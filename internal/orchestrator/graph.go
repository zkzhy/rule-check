package orchestrator

import (
	"context"
	"fmt"

	"audit-workflow/internal/components/tools/submit"
	"audit-workflow/internal/config"
	"audit-workflow/internal/fetch"

	"github.com/cloudwego/eino/compose"
)

type WorkflowInput struct{}

type WorkflowOutput struct{}

func BuildWorkflow(ctx context.Context, cfg *config.RootConfig) (compose.Runnable[WorkflowInput, WorkflowOutput], error) {
	graph := compose.NewGraph[WorkflowInput, WorkflowOutput]()

	fetchNode := compose.InvokableLambda(func(ctx context.Context, in WorkflowInput) (WorkflowInput, error) {
		if err := fetch.Run(cfg); err != nil {
			return in, fmt.Errorf("fetch failed: %w", err)
		}
		return in, nil
	})
	if err := graph.AddLambdaNode("fetch", fetchNode); err != nil {
		return nil, err
	}

	aiNode := compose.InvokableLambda(func(ctx context.Context, in WorkflowInput) (WorkflowInput, error) {
		if err := RunRiskAnalysis(ctx, cfg); err != nil {
			return in, fmt.Errorf("ai failed: %w", err)
		}
		return in, nil
	})
	if err := graph.AddLambdaNode("ai", aiNode); err != nil {
		return nil, err
	}

	submitNode := compose.InvokableLambda(func(ctx context.Context, in WorkflowInput) (WorkflowOutput, error) {
		if err := submit.Run(cfg); err != nil {
			return WorkflowOutput{}, fmt.Errorf("submit failed: %w", err)
		}
		return WorkflowOutput{}, nil
	})
	if err := graph.AddLambdaNode("submit", submitNode); err != nil {
		return nil, err
	}

	if err := graph.AddEdge(compose.START, "fetch"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("fetch", "ai"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("ai", "submit"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("submit", compose.END); err != nil {
		return nil, err
	}

	compiled, err := graph.Compile(ctx)
	if err != nil {
		return nil, err
	}
	return compiled, nil
}
