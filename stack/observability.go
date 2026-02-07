package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const otelCollectorReleaseName = "otel-collector"

func baseHelmValues() pulumi.Map {
	return pulumi.Map{
		"fullnameOverride": pulumi.String(trinoReleaseName),
		"additionalConfigProperties": pulumi.Array{
			pulumi.String("log.path=tcp://" + otelCollectorReleaseName + ":54525"),
			pulumi.String("log.format=json"),
			pulumi.String("tracing.enabled=true"),
			pulumi.String("tracing.exporter.endpoint=http://" + otelCollectorReleaseName + ":4317"),
		},
	}
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
								"metrics_path":    pulumi.String("/metrics"),
								"static_configs": pulumi.Array{
									pulumi.Map{
										"targets": pulumi.Array{
											pulumi.String(trinoReleaseName + ":8080"),
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
