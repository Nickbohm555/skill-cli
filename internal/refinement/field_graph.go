package refinement

import (
	"fmt"
	"sort"
)

// FieldGraph stores direct dependency edges and exposes stable transitive
// impact queries for revision-aware revalidation.
type FieldGraph struct {
	order      map[FieldID]int
	dependents map[FieldID][]FieldID
}

// DefaultFieldGraph returns the phase-3 dependency graph in deterministic order.
func DefaultFieldGraph() FieldGraph {
	graph, err := NewFieldGraph(
		DefaultFieldRegistry(),
		map[FieldID][]FieldID{
			FieldPurposeSummary: {
				FieldPrimaryTasks,
				FieldSuccessCriteria,
				FieldExampleRequests,
				FieldInScope,
				FieldOutOfScope,
			},
			FieldPrimaryTasks: {
				FieldExampleRequests,
				FieldExampleOutputs,
			},
			FieldSuccessCriteria: {
				FieldExampleOutputs,
			},
			FieldConstraints: {
				FieldDependencies,
				FieldExampleOutputs,
			},
			FieldDependencies: {
				FieldExampleOutputs,
			},
			FieldExampleRequests: {
				FieldExampleOutputs,
			},
			FieldInScope: {
				FieldExampleRequests,
				FieldOutOfScope,
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return graph
}

func NewFieldGraph(registry []FieldDefinition, directDependents map[FieldID][]FieldID) (FieldGraph, error) {
	if len(registry) == 0 {
		return FieldGraph{}, fmt.Errorf("field graph registry cannot be empty")
	}

	order := make(map[FieldID]int, len(registry))
	allowed := make(map[FieldID]struct{}, len(registry))
	for index, def := range registry {
		order[def.ID] = index
		allowed[def.ID] = struct{}{}
	}

	graph := FieldGraph{
		order:      order,
		dependents: make(map[FieldID][]FieldID, len(registry)),
	}

	for _, def := range registry {
		graph.dependents[def.ID] = nil
	}

	for source, dependents := range directDependents {
		if _, ok := allowed[source]; !ok {
			return FieldGraph{}, fmt.Errorf("field graph source %q is not registered", source)
		}

		seen := make(map[FieldID]struct{}, len(dependents))
		normalized := make([]FieldID, 0, len(dependents))
		for _, dependent := range dependents {
			if _, ok := allowed[dependent]; !ok {
				return FieldGraph{}, fmt.Errorf("field graph dependent %q is not registered", dependent)
			}
			if dependent == source {
				return FieldGraph{}, fmt.Errorf("field graph %q cannot depend on itself", source)
			}
			if _, exists := seen[dependent]; exists {
				continue
			}
			seen[dependent] = struct{}{}
			normalized = append(normalized, dependent)
		}

		sort.Slice(normalized, func(i, j int) bool {
			return graph.order[normalized[i]] < graph.order[normalized[j]]
		})
		graph.dependents[source] = normalized
	}

	return graph, nil
}

func (g FieldGraph) DirectDependents(fieldID FieldID) []FieldID {
	return append([]FieldID(nil), g.dependents[fieldID]...)
}

func (g FieldGraph) ImpactedBy(fieldID FieldID) []FieldID {
	seen := make(map[FieldID]struct{})
	queue := append([]FieldID(nil), g.dependents[fieldID]...)

	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]

		if _, visited := seen[next]; visited {
			continue
		}
		seen[next] = struct{}{}
		queue = append(queue, g.dependents[next]...)
	}

	impacted := make([]FieldID, 0, len(seen))
	for fieldID := range seen {
		impacted = append(impacted, fieldID)
	}

	sort.Slice(impacted, func(i, j int) bool {
		return g.order[impacted[i]] < g.order[impacted[j]]
	})

	return impacted
}
