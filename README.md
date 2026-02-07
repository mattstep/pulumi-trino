# pulumi-trino

A [Pulumi](https://www.pulumi.com/) project that deploys a [Trino](https://trino.io) cluster on Kubernetes using the official [Trino Helm chart](https://trino.io/docs/current/installation/kubernetes.html), with an [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) for observability.

Supports two deployment targets via Go build tags:

- **EKS** (default) -- provisions a VPC and EKS cluster on AWS, exposes Trino via LoadBalancer
- **Minikube** -- uses a local minikube cluster with resource-constrained settings

## What gets deployed

- **Trino** (Helm chart v1.42.0) -- coordinator (and workers on EKS)
- **OpenTelemetry Collector** (contrib image) -- receives:
  - **Metrics** -- scrapes Trino's native `/metrics` endpoint via Prometheus receiver
  - **Logs** -- receives Trino's structured JSON logs via TCP
  - **Traces** -- receives Trino's OpenTelemetry traces via OTLP gRPC

## Prerequisites

- [Go](https://go.dev/doc/install)
- [Pulumi CLI](https://www.pulumi.com/docs/install/)

For EKS:
- [AWS credentials](https://www.pulumi.com/registry/packages/aws/installation-configuration/#set-credentials-as-environment-variables) configured in your environment

For Minikube:
- [minikube](https://minikube.sigs.k8s.io/docs/start/) installed and running

## Deploying to EKS

1. [Login to Pulumi](https://www.pulumi.com/docs/cli/commands/pulumi_login/)
2. Deploy from the `stack/` directory:
   ```
   cd stack
   pulumi up
   ```

Alternatively, use the Automation API entrypoint at the repo root to deploy remotely via Pulumi Deployments (update `main.go` to point to your organization first):
```
go run main.go
```

To destroy the stack:
```
cd stack
pulumi destroy
```

## Deploying to Minikube

1. Start minikube:
   ```
   minikube start --cpus=2 --memory=4096
   ```

2. Deploy with the `minikube` build tag:
   ```
   cd stack
   GOFLAGS="-tags=minikube" pulumi login --local
   GOFLAGS="-tags=minikube" pulumi stack init dev
   GOFLAGS="-tags=minikube" pulumi up --yes
   ```

   The minikube build reads your local kubeconfig (`~/.kube/config` by default) and configures Trino with reduced resource limits (512M coordinator heap, no workers) to fit within a 4GB minikube VM.

3. Verify Trino is running:
   ```
   SVC_NAME=$(kubectl get svc -l app.kubernetes.io/name=trino -o jsonpath='{.items[0].metadata.name}')
   kubectl port-forward "svc/$SVC_NAME" 8080:8080 &
   curl http://localhost:8080/v1/info
   ```

To destroy:
```
GOFLAGS="-tags=minikube" pulumi destroy --yes
```

## Project structure

```
stack/
  main.go           -- entrypoint, calls deploy()
  stack.go          -- orchestration: creates cluster, deploys OTel collector and Trino
  observability.go  -- OpenTelemetry Collector Helm release and configuration
  eks.go            -- EKS cluster provisioning and Helm values (build tag: !minikube)
  minikube.go       -- local kubeconfig reader and Helm values (build tag: minikube)
  Pulumi.yaml       -- Pulumi project definition
main.go             -- Automation API entrypoint for remote EKS deployment
```

## CI

A GitHub Actions workflow (`.github/workflows/test-minikube.yml`) runs on pushes and PRs to `main`. It deploys to minikube and verifies that Trino's `/v1/info` endpoint returns a 2xx response and the OTel collector is running.
