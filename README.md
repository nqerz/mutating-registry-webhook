# Mutating Registry Webhook

A Kubernetes mutating webhook that automatically transforms container image registry paths in pod specifications, enabling seamless registry redirection and mirroring capabilities.

## Overview

This webhook service intercepts Kubernetes pod creation and update requests, dynamically modifying container image registries based on configured mapping rules. This functionality is essential for organizations that need to:

- Redirect container images through private registry mirrors
- Enforce corporate policy by routing public registry requests to approved internal mirrors
- Implement high-availability registry failover mechanisms
- Maintain consistent image sourcing across multiple environments

## Features

- **Dynamic Registry Mapping**: Configure registry transformations through a Kubernetes ConfigMap
- **Secure Communication**: TLS-encrypted webhook endpoints for secure admission control
- **Flexible Deployment**: Helm-based deployment for easy installation and upgrades
- **Configurable Rules**: Fine-grained control over registry transformation patterns
- **Kubernetes Native**: Seamlessly integrates with Kubernetes admission control system
- **Resource Efficient**: Lightweight implementation with configurable resource limits

## Prerequisites

Before installing the webhook, ensure you have:

- Kubernetes cluster (version 1.16 or higher)
- Helm 3.0 or higher installed
- kubectl configured with access to your cluster
- Cluster admin privileges for webhook configuration

## Installation

### 1. Generate TLS Certificates

First, generate the required TLS certificates for secure webhook communication:

```bash
make generate-certs
```

### 2. Deploy with Helm

Install the webhook using Helm:

```bash
make deploy
```

Verify the deployment:

```bash
make verify-deployment
```

## Configuration

### Registry Mappings

Configure registry mappings in the `values.yaml` file:

```yaml
registryMappings:
  "docker.io": "mirror.example.com"
  "gcr.io": "gcr.mirror.example.com"
```

## Development

### Building

Build the webhook binary:

```bash
make build
```

Build and push Docker image:

```bash
make docker-build docker-push
```

### Testing

Run the test suite:

```bash
make test
```

### Deployment for Development

Deploy the webhook to a development cluster:

```bash
make test-deploy
```

Verify the deployment:

```bash
make verify-deployment
```

### Cleanup

To clean up build artifacts:

```bash
make clean
