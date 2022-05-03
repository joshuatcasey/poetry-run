package poetryrun

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface PyProjectParser --output fakes/py_project_parser.go

// BuildPlanMetadata is the buildpack specific data included in build plan
// requirements.
type BuildPlanMetadata struct {
	// Build denotes the dependency is needed at build-time.
	Launch bool `toml:"launch"`
}

type PyProjectParser interface {
	Parse(string) (string, error)
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection will contribute a Build Plan that provides site-packages,
// and requires cpython and pip at build.
//
// Detection is contingent on there being one or more scripts to run
// defined in the pyproject.toml under [tool.poetry.scripts]
func Detect(pyProjectParser PyProjectParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		script, err := pyProjectParser.Parse(filepath.Join(context.WorkingDir, "pyproject.toml"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if script == "" {
			return packit.DetectResult{}, packit.Fail
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: CPython,
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			},
			{
				Name: Poetry,
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			},
			{
				Name: PoetryVenv,
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			},
		}

		if shouldReload, err := checkLiveReloadEnabled(); err != nil {
			return packit.DetectResult{}, err
		} else if shouldReload {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Watchexec,
				Metadata: BuildPlanMetadata{
					Launch: true,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{},
				Requires: requirements,
			},
		}, nil
	}
}

func checkLiveReloadEnabled() (bool, error) {
	if reload, ok := os.LookupEnv("BP_LIVE_RELOAD_ENABLED"); ok {
		shouldEnableReload, err := strconv.ParseBool(reload)
		if err != nil {
			return false, packit.Fail.WithMessage("failed to parse BP_LIVE_RELOAD_ENABLED value %s: %w", reload, err)
		}
		return shouldEnableReload, nil
	}
	return false, nil
}
