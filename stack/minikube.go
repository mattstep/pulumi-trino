//go:build minikube

package main

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createCluster(ctx *pulumi.Context) (pulumi.StringOutput, []pulumi.Resource, error) {
	conf := config.New(ctx, "")
	kubeconfigPath := conf.Get("kubeconfig")
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return pulumi.StringOutput{}, nil, fmt.Errorf("failed to determine home directory: %w", err)
		}
		kubeconfigPath = home + "/.kube/config"
	}

	kubeconfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return pulumi.StringOutput{}, nil, fmt.Errorf("failed to read kubeconfig from %s: %w", kubeconfigPath, err)
	}

	return pulumi.String(string(kubeconfig)).ToStringOutput(), nil, nil
}

func helmValues() pulumi.Map {
	return pulumi.Map{
		"service": pulumi.Map{
			"type": pulumi.String("ClusterIP"),
		},
		"server": pulumi.Map{
			"workers": pulumi.Int(0),
			"config": pulumi.Map{
				"query": pulumi.Map{
					"maxMemory": pulumi.String("512MB"),
				},
			},
		},
		"coordinator": pulumi.Map{
			"jvm": pulumi.Map{
				"maxHeapSize": pulumi.String("1G"),
			},
			"config": pulumi.Map{
				"nodeScheduler": pulumi.Map{
					"includeCoordinator": pulumi.Bool(true),
				},
				"query": pulumi.Map{
					"maxMemoryPerNode": pulumi.String("256MB"),
				},
			},
		},
	}
}
