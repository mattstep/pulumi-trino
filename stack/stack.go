package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	otelCollectorReleaseName = "otel-collector"
	trinoReleaseName         = "example-trino-cluster"
)

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

func installOtelCollector(ctx *pulumi.Context, provider *kubernetes.Provider) (*helm.Release, error) {
	return helm.NewRelease(ctx,
		otelCollectorReleaseName,
		&helm.ReleaseArgs{
			Name:  pulumi.String(otelCollectorReleaseName),
			Chart: pulumi.String("opentelemetry-collector"),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://open-telemetry.github.io/opentelemetry-helm-charts"),
			},
			Values: otelCollectorValues(),
		},
		pulumi.Provider(provider))
}

func otelCollectorValues() pulumi.Map {
	return pulumi.Map{
		"fullnameOverride": pulumi.String(otelCollectorReleaseName),
		"mode":             pulumi.String("deployment"),
		"image": pulumi.Map{
			"repository": pulumi.String("otel/opentelemetry-collector-contrib"),
		},
		"resources": pulumi.Map{
			"limits": pulumi.Map{
				"memory": pulumi.String("128Mi"),
			},
			"requests": pulumi.Map{
				"cpu":    pulumi.String("50m"),
				"memory": pulumi.String("64Mi"),
			},
		},
		"config": pulumi.Map{
			"receivers": pulumi.Map{
				"otlp": pulumi.Map{
					"protocols": pulumi.Map{
						"grpc": pulumi.Map{
							"endpoint": pulumi.String("${env:MY_POD_IP}:4317"),
						},
					},
				},
				"prometheus": pulumi.Map{
					"config": pulumi.Map{
						"scrape_configs": pulumi.Array{
							pulumi.Map{
								"job_name":        pulumi.String("trino"),
								"scrape_interval": pulumi.String("30s"),
								"static_configs": pulumi.Array{
									pulumi.Map{
										"targets": pulumi.Array{
											pulumi.String(trinoReleaseName + ":5556"),
										},
									},
								},
							},
						},
					},
				},
				"tcplog": pulumi.Map{
					"listen_address": pulumi.String("0.0.0.0:54525"),
				},
			},
			"processors": pulumi.Map{
				"batch": pulumi.Map{},
				"memory_limiter": pulumi.Map{
					"check_interval":  pulumi.String("5s"),
					"limit_mib":       pulumi.Int(100),
					"spike_limit_mib": pulumi.Int(25),
				},
			},
			"exporters": pulumi.Map{
				"debug": pulumi.Map{},
			},
			"extensions": pulumi.Map{
				"health_check": pulumi.Map{
					"endpoint": pulumi.String("${env:MY_POD_IP}:13133"),
				},
			},
			"service": pulumi.Map{
				"extensions": pulumi.Array{
					pulumi.String("health_check"),
				},
				"pipelines": pulumi.Map{
					"metrics": pulumi.Map{
						"receivers":  pulumi.Array{pulumi.String("prometheus")},
						"processors": pulumi.Array{pulumi.String("memory_limiter"), pulumi.String("batch")},
						"exporters":  pulumi.Array{pulumi.String("debug")},
					},
					"logs": pulumi.Map{
						"receivers":  pulumi.Array{pulumi.String("tcplog")},
						"processors": pulumi.Array{pulumi.String("memory_limiter"), pulumi.String("batch")},
						"exporters":  pulumi.Array{pulumi.String("debug")},
					},
					"traces": pulumi.Map{
						"receivers":  pulumi.Array{pulumi.String("otlp")},
						"processors": pulumi.Array{pulumi.String("memory_limiter"), pulumi.String("batch")},
						"exporters":  pulumi.Array{pulumi.String("debug")},
					},
				},
			},
		},
		"ports": pulumi.Map{
			"otlp": pulumi.Map{
				"enabled":       pulumi.Bool(true),
				"containerPort": pulumi.Int(4317),
				"servicePort":   pulumi.Int(4317),
				"protocol":      pulumi.String("TCP"),
				"appProtocol":   pulumi.String("grpc"),
			},
			"tcplog": pulumi.Map{
				"enabled":       pulumi.Bool(true),
				"containerPort": pulumi.Int(54525),
				"servicePort":   pulumi.Int(54525),
				"protocol":      pulumi.String("TCP"),
			},
		},
	}
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
