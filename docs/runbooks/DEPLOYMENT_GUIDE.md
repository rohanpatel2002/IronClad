# Production Deployment & Infrastructure Runbook

## Overview
This runbook provides step-by-step instructions for provisioning, configuring, and operating the IRONCLAD platform infrastructure on AWS EKS and Kubernetes.

## Prerequisites
- Terraform >= 1.5.0
- Helm >= 3.12
- kubectl configured with EKS admin credentials

## Provisioning Infrastructure
```bash
cd infra/terraform
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

## Helm Deployment
```bash
helm upgrade --install ironclad infra/helm/ironclad -n ironclad --create-namespace
```
