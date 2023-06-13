package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremoteup"
)

func main() {
	ctx := context.Background()

	stackName := auto.FullyQualifiedStackName("mattstep", "pulumi-trino", "demo")

	repo := auto.GitRepo{
		URL:         "https://github.com/mattstep/pulumi-trino.git",
		Branch:      "refs/heads/main",
		ProjectPath: "/stack",
	}

	env := map[string]auto.EnvVarValue{
		"AWS_REGION":            {Value: "us-west-2"},
		"AWS_ACCESS_KEY_ID":     {Value: os.Getenv("AWS_ACCESS_KEY_ID")},
		"AWS_SECRET_ACCESS_KEY": {Value: os.Getenv("AWS_SECRET_ACCESS_KEY"), Secret: true},
		"AWS_SESSION_TOKEN":     {Value: os.Getenv("AWS_SESSION_TOKEN"), Secret: true},
	}

	stack, err := auto.UpsertRemoteStackGitSource(ctx, stackName, repo, auto.RemoteEnvVars(env))
	if err != nil {
		fmt.Printf("Failed to create or select stack: %v\n", err)
		os.Exit(1)
	}

	stdoutStreamer := optremoteup.ProgressStreams(os.Stdout)
	_, err = stack.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to update stack: %v\n\n", err)
		os.Exit(1)
	}
}