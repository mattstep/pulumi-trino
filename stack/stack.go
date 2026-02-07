package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ClusterInfo holds the information needed to connect to a Kubernetes cluster
// and deploy workloads, regardless of how the cluster was provisioned.
type ClusterInfo struct {
	Kubeconfig pulumi.StringOutput
	DependsOn  []pulumi.Resource
	HelmValues pulumi.Map
}

func deploy() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clusterInfo, err := createCluster(ctx)
		if err != nil {
			return err
		}

		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetesProvider", &kubernetes.ProviderArgs{
			Kubeconfig: clusterInfo.Kubeconfig,
		}, pulumi.DependsOn(clusterInfo.DependsOn))
		if err != nil {
			return err
		}

		return installTrinoHelmChart(ctx, kubeProvider, clusterInfo.HelmValues)
	})
}

func installTrinoHelmChart(ctx *pulumi.Context, provider *kubernetes.Provider, values pulumi.Map) error {
	_, err := helm.NewRelease(ctx,
		"example-trino-cluster",
		&helm.ReleaseArgs{
			Name:    pulumi.String("example-trino-cluster"),
			Chart:   pulumi.String("trino"),
			Version: pulumi.String("1.42.0"),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://trinodb.github.io/charts"),
			},
			Values:      values,
			WaitForJobs: pulumi.Bool(true),
		},
		pulumi.Provider(provider))

	return err
}
