package esc_core

import (
	"fmt"
	"sort"
	"strings"
)

type Runner struct {
	stages  []StageID
	systems []System
}

func NewRunner(stages []StageID) *Runner {
	return &Runner{
		stages:  stages,
		systems: make([]System, 0, len(stages)),
	}
}

func (r *Runner) Add(system System) {
	r.systems = append(r.systems, system)
}

func (r *Runner) ValidateAccess() error {
	var accessByStage = make(map[StageID]AccessSet, len(r.stages))
	var ownerByStage = make(map[StageID]map[ComponentID]string, len(r.stages))

	for _, stage := range r.stages {
		accessByStage[stage] = NewAccessSet()
		ownerByStage[stage] = make(map[ComponentID]string)
	}

	for _, system := range r.systems {
		var stage = system.Stage()

		var stageAccess, ok = accessByStage[stage]

		if !ok {
			return fmt.Errorf("unknown stage %d for system %s", stage, system.Name())
		}

		var conflicts = stageAccess.ConflictsWith(system.Access())

		if len(conflicts) > 0 {
			var components = conflicts[0]

			return AccessConflict{
				Stage:     stage,
				Component: components,
				First:     ownerByStage[stage][components],
				Second:    system.Name(),
			}
		}

		stageAccess.Merge(system.Access())

		for component := range system.Access().Reads {
			if _, exists := ownerByStage[stage][component]; !exists {
				ownerByStage[stage][component] = system.Name()
			}
		}

		for component := range system.Access().Writes {
			ownerByStage[stage][component] = system.Name()
		}
	}

	return nil
}

func (r *Runner) Update(ctx *Context) error {
	for _, stage := range r.stages {
		ctx.Stage = stage

		ctx.World.mutationPhase = MutationRunningSystem

		for _, system := range r.systems {
			if system.Stage() != stage {
				continue
			}

			var err = system.Update(ctx)

			if err != nil {
				return err
			}
		}

		ctx.World.mutationPhase = MutationApplyingCommands

		var err = ctx.Commands.Apply(ctx.World)

		if err != nil {
			return err
		}

		ctx.Commands.Clear()

		ctx.World.mutationPhase = MutationIdle
	}

	return nil
}

func (r *Runner) InspectAccess() AccessReport {
	report := AccessReport{
		Stages: make([]StageAccessReport, 0, len(r.stages)),
	}

	for _, stage := range r.stages {
		stageReport := StageAccessReport{
			Stage:   stage,
			Systems: make([]SystemAccessInfo, 0),
		}

		for _, system := range r.systems {
			if system.Stage() != stage {
				continue
			}

			queries := system.DebugQueries()
			normalizedQueries := make([]QueryDebugInfo, 0, len(queries))

			for _, query := range queries {
				if query.System == "" {
					query.System = system.Name()
				}

				normalizedQueries = append(normalizedQueries, query)
			}

			stageReport.Systems = append(stageReport.Systems, SystemAccessInfo{
				System:  system.Name(),
				Reads:   collectQueryComponents(normalizedQueries, true),
				Writes:  collectQueryComponents(normalizedQueries, false),
				Queries: normalizedQueries,
			})
		}

		report.Stages = append(report.Stages, stageReport)
	}

	return report
}

func (r *Runner) DebugAccess() string {
	var report = r.InspectAccess()

	var builder strings.Builder

	_, _ = fmt.Fprintf(&builder, "Access:\\n")

	for _, stage := range report.Stages {
		_, _ = fmt.Fprintf(&builder, "\nStage %d\n", stage.Stage)

		if len(stage.Systems) == 0 {
			_, _ = fmt.Fprintf(&builder, "  systems: 0\n")
			continue
		}

		for _, system := range stage.Systems {
			_, _ = fmt.Fprintf(&builder, "  %s\n", system.System)
			_, _ = fmt.Fprintf(&builder, "    reads: %s\n", formatDebugList(system.Reads))
			_, _ = fmt.Fprintf(&builder, "    writes: %s\n", formatDebugList(system.Writes))
		}
	}

	return builder.String()
}

func (r *Runner) DebugQueries() string {
	var report = r.InspectAccess()
	var builder strings.Builder

	_, _ = fmt.Fprintf(&builder, "Queries:\n")

	for _, stage := range report.Stages {
		_, _ = fmt.Fprintf(&builder, "\nStage %d\n", stage.Stage)

		if len(stage.Systems) == 0 {
			_, _ = fmt.Fprintf(&builder, "  systems: 0\n")
			continue
		}

		for _, system := range stage.Systems {
			_, _ = fmt.Fprintf(&builder, "  %s\n", system.System)

			if len(system.Queries) == 0 {
				_, _ = fmt.Fprintf(&builder, "  queries: 0\n")
				continue
			}

			for _, query := range system.Queries {
				_, _ = fmt.Fprintf(&builder, "    %s\n", query.Query)
				_, _ = fmt.Fprintf(&builder, "      with: %s\n", formatDebugList(query.With))
				_, _ = fmt.Fprintf(&builder, "      without: %s\n", formatDebugList(query.Without))
				_, _ = fmt.Fprintf(&builder, "      reads: %s\n", formatDebugList(query.Reads))
				_, _ = fmt.Fprintf(&builder, "      writes: %s\n", formatDebugList(query.Writes))
			}
		}
	}

	return builder.String()
}

func formatDebugList(list []string) string {
	if len(list) == 0 {
		return "-"
	}

	var copied = append([]string(nil), list...)

	sort.Strings(copied)

	return strings.Join(copied, ", ")
}

func collectQueryComponents(queries []QueryDebugInfo, reads bool) []string {
	seen := make(map[string]struct{})

	for _, query := range queries {
		var components []string
		if reads {
			components = query.Reads
		} else {
			components = query.Writes
		}

		for _, component := range components {
			seen[component] = struct{}{}
		}
	}

	result := make([]string, 0, len(seen))
	for component := range seen {
		result = append(result, component)
	}

	sort.Strings(result)
	return result
}
