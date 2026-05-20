package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"refactorlah/internal/planning"
)

var ErrAdapterFailure = errors.New("adapter failure")

type Invoker struct{}

func NewInvoker() *Invoker {
	return &Invoker{}
}

func (i *Invoker) Invoke(ctx context.Context, projectRoot string, plan planning.MovePlan, dryRun bool, selection Selection) (AggregatedResponse, error) {
	aggregate := AggregatedResponse{}
	for _, adapter := range selection.Adapters {
		request := Request{
			ProtocolVersion: 1,
			ProjectRoot:     ".",
			OldPath:         plan.OldPath,
			NewPath:         plan.NewPath,
			DryRun:          dryRun,
			Options:         adapter.Options,
		}

		for _, move := range plan.Moves {
			request.Moves = append(request.Moves, Move{
				OldPath: move.OldPath,
				NewPath: move.NewPath,
				Tracked: move.Tracked,
			})
		}

		payload, err := json.Marshal(request)
		if err != nil {
			return AggregatedResponse{}, err
		}

		command := exec.CommandContext(ctx, adapter.Path, "analyze")
		command.Dir = projectRoot
		command.Stdin = bytes.NewReader(payload)

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr

		if err := command.Run(); err != nil {
			return AggregatedResponse{}, fmt.Errorf("%w: %s: %v: %s", ErrAdapterFailure, adapter.Name, err, strings.TrimSpace(stderr.String()))
		}

		var response Response
		if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
			return AggregatedResponse{}, fmt.Errorf("%w: %s returned invalid JSON: %v", ErrAdapterFailure, adapter.Name, err)
		}
		if response.ProtocolVersion != 1 {
			return AggregatedResponse{}, fmt.Errorf("%w: %s returned unsupported protocol version %d", ErrAdapterFailure, adapter.Name, response.ProtocolVersion)
		}
		if len(response.Errors) > 0 {
			return AggregatedResponse{}, fmt.Errorf("%w: %s: %s", ErrAdapterFailure, adapter.Name, strings.Join(response.Errors, "; "))
		}

		for index := range response.Replacements {
			response.Replacements[index].Adapter = adapter.Name
		}

		aggregate.SymbolMappings = append(aggregate.SymbolMappings, response.SymbolMappings...)
		aggregate.PathMappings = append(aggregate.PathMappings, response.PathMappings...)
		aggregate.Replacements = append(aggregate.Replacements, response.Replacements...)
		aggregate.Warnings = append(aggregate.Warnings, response.Warnings...)
	}

	return aggregate, nil
}
