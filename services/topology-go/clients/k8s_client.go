package clients

import (
	"context"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	AnnotationDependsOn   = "ironclad.security/depends-on"
	AnnotationCriticality = "ironclad.security/criticality"
)

// K8sClient wraps the Kubernetes client-go library
type K8sClient struct {
	clientset *kubernetes.Clientset
}

// NewK8sClient creates a new Kubernetes client
func NewK8sClient() (*K8sClient, error) {
	var config *rest.Config
	var err error

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		// Use out-of-cluster config
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// Fallback to in-cluster config
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}

	return &K8sClient{clientset: clientset}, nil
}

// ServiceMetadata represents extracted topology information for a service
type ServiceMetadata struct {
	Name        string
	DependsOn   []string
	Criticality float64
}

// GetServiceTopology queries K8s for Services across all namespaces and
// extracts the dependency graph from IRONCLAD annotations.
func (c *K8sClient) GetServiceTopology(ctx context.Context) ([]ServiceMetadata, error) {
	// Query Services in all namespaces
	servicesList, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list k8s services: %w", err)
	}

	var metadata []ServiceMetadata

	for _, svc := range servicesList.Items {
		// We only care about services that have IronClad annotations or we can parse them generally
		dependsOnStr := svc.Annotations[AnnotationDependsOn]
		
		var dependsOn []string
		if dependsOnStr != "" {
			for _, dep := range strings.Split(dependsOnStr, ",") {
				cleanDep := strings.TrimSpace(dep)
				if cleanDep != "" {
					dependsOn = append(dependsOn, cleanDep)
				}
			}
		}

		// Parse criticality if available, default to 0.5
		criticality := 0.5
		if critStr, ok := svc.Annotations[AnnotationCriticality]; ok {
			var parsed float64
			if _, err := fmt.Sscanf(critStr, "%f", &parsed); err == nil {
				criticality = parsed
			}
		}

		metadata = append(metadata, ServiceMetadata{
			Name:        svc.Name,
			DependsOn:   dependsOn,
			Criticality: criticality,
		})
	}

	return metadata, nil
}
