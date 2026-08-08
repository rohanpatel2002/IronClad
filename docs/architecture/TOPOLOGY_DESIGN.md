# Topology Service Design & K8s Sync Model

The `topology-go` service maintains an in-memory directed acyclic graph (DAG) representing service-to-service dependencies across the microservices ecosystem.

## Core Features
1. **Dynamic K8s Graph Builder**: Automatically syncs service labels and annotations from Kubernetes deployments.
2. **BFS Blast Radius Traversal**: Evaluates upstream impact (services that call into the modified service) and direct downstream dependencies.
3. **Cycle Detection**: Identifies circular dependencies in the service graph.
4. **SLO Error Budget Monitoring**: Tracks availability objectives and error budget consumption per service node.

## Graph Schema
```json
{
  "name": "payment-api",
  "criticality": 1.00,
  "depends_on": ["auth-service", "fraud-detection", "notification-service", "audit-logger"],
  "depended_on_by": ["api-gateway", "order-service"]
}
```
