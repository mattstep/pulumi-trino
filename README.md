# pulumi-trino
A simple example for using [Pulumi](https://www.pulumi.com/) to provision an AWS EKS cluster with a simple [Trino](https://trino.io) cluster deployed via the official [Trino Helm Chart](https://trino.io/docs/current/installation/kubernetes.html]).

## prerequisites
- [Golang](https://go.dev/doc/install)
- [Pulumi CLI](https://www.pulumi.com/docs/install/)

## usage
1. Clone this repo
2. Run `go get` to download dependencies
3. Ensure you have [AWS Credentials](https://www.pulumi.com/registry/packages/aws/installation-configuration/#set-credentials-as-environment-variables) in your environment
4. [Login to Pulumi](https://www.pulumi.com/docs/cli/commands/pulumi_login/)
5. Update `main.go` to point to your organization (it's currently pointed to mine, `mattstep`)
6. Run `go run main.go` from the root

To destroy the stack later, run `cd stack` followed by `pulumi destroy`.
