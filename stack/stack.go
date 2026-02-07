package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const trinoReleaseName = "example-trino-cluster"

func deploy() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		kubeconfig, dependsOn, err := createCluster(ctx)
		if err != nil {
			return err
		}

		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetesProvider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeconfig,
		}, pulumi.DependsOn(dependsOn))
		if err != nil {
			return err
		}

		otelCollector, err := installOtelCollector(ctx, kubeProvider)
		if err != nil {
			return err
		}

		return installTrinoHelmChart(ctx, kubeProvider, helmValues(), otelCollector)
	})
}

func installTrinoHelmChart(ctx *pulumi.Context, provider *kubernetes.Provider, values pulumi.Map, otelCollector *helm.Release) error {
	_, err := helm.NewRelease(ctx,
		trinoReleaseName,
		&helm.ReleaseArgs{
			Name:    pulumi.String(trinoReleaseName),
			Chart:   pulumi.String("trino"),
			Version: pulumi.String("1.42.0"),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://trinodb.github.io/charts"),
			},
			Values:      values,
			WaitForJobs: pulumi.Bool(true),
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{otelCollector}))

	return err
}
