package main

import (
	"github.com/pulumi/pulumi-awsx/sdk/go/awsx/ec2"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createEks(ctx *pulumi.Context) (*eks.Cluster, error) {
	var eksVpc, err = ec2.NewVpc(ctx, "trino-vpc", &ec2.VpcArgs{
		EnableDnsHostnames: pulumi.Bool(true),
	})

	cluster, err := eks.NewCluster(ctx, "trino", &eks.ClusterArgs{
		VpcId:            eksVpc.VpcId,
		PublicSubnetIds:  eksVpc.PublicSubnetIds,
		PrivateSubnetIds: eksVpc.PrivateSubnetIds,
		InstanceType:     pulumi.String("r6a.large"),
		DesiredCapacity:  pulumi.Int(3),
		MinSize:          pulumi.Int(3),
		MaxSize:          pulumi.Int(10),
	})

	if err != nil {
		return nil, err
	}
	ctx.Export("kubeconfig", cluster.Kubeconfig)
	return cluster, nil
}

func installTrinoHelmChart(ctx *pulumi.Context, provider *kubernetes.Provider) error {
	_, err := helm.NewChart(ctx,
		"example-trino-cluster",
		helm.ChartArgs{
			Chart:     pulumi.String("trino"),
			Version:   pulumi.String("0.10.2"),
			FetchArgs: helm.FetchArgs{
				Repo: pulumi.String("https://trinodb.github.io/charts"),
			},
			Values: pulumi.Map{
				"service": pulumi.Map{
					"type": pulumi.String("LoadBalancer"),
					"port": pulumi.String("80"),
				},
			},
		},
		pulumi.Provider(provider))

	if err != nil {
		return err
	}
	return nil
}

func deploy() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cluster, err := createEks(ctx)
		if err != nil {
			return err
		}

		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetesProvider", &kubernetes.ProviderArgs{
			Kubeconfig: cluster.KubeconfigJson,
		}, pulumi.DependsOn([]pulumi.Resource{cluster}))
		if err != nil {
			return err
		}

		err = installTrinoHelmChart(ctx, kubeProvider)
		return err
	})
}
