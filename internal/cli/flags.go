package cli

import (
	"strings"

	"github.com/respawn-app/ksrc/internal/gradle"
)

type ResolveFlags struct {
	Project         string
	Module          string
	Group           string
	Artifact        string
	Version         string
	Scope           string
	Config          string
	Targets         string
	Subprojects     []string
	Offline         bool
	Refresh         bool
	All             bool
	IncludeBuildSrc bool
}

func (f ResolveFlags) ToOptions() gradle.ResolveOptions {
	return gradle.ResolveOptions{
		ProjectDir:      f.Project,
		Module:          f.Module,
		Group:           f.Group,
		Artifact:        f.Artifact,
		Version:         f.Version,
		Scope:           f.Scope,
		Configs:         splitCSV(f.Config),
		Targets:         splitCSV(f.Targets),
		Subprojects:     f.Subprojects,
		Offline:         f.Offline,
		Refresh:         f.Refresh,
		IncludeBuildSrc: f.IncludeBuildSrc,
	}
}

func splitCSV(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
