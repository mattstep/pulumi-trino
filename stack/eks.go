//go:build !minikube

package main

import (
	"github.com/pulumi/pulumi-awsx/sdk/go/awsx/ec2"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCluster(ctx *pulumi.Context) (*ClusterInfo, error) {
	eksVpc, err := ec2.NewVpc(ctx, "trino-vpc", &ec2.VpcArgs{
		EnableDnsHostnames: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

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

	return &ClusterInfo{
		Kubeconfig: cluster.KubeconfigJson,
		DependsOn:  []pulumi.Resource{cluster},
		HelmValues: pulumi.Map{
			"service": pulumi.Map{
				"type": pulumi.String("LoadBalancer"),
			},
		},
	}, nil
}
