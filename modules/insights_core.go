package modules

import (
	"fmt"
	"os"
)

func GetAdvisorModule() *Module {
	return &Module{
		Name:    "advisor",
		Version: "??? (not implemented)",
		Env:     getInsightsCoreEnv(),
		Exec:    []string{"python3", "-m", "insights.client.phase.v2"},
		Commands: [][]string{
			{"advisor"},
			{"advisor", "collect"},
			{"advisor", "check-results"},
			{"advisor", "show-results"},
			{"advisor", "list-specs"},
			{"advisor", "diagnosis"},
		},
		CollectCommand:     []string{"advisor", "collect"},
		ArchiveContentType: "application/vnd.redhat.advisor.collection",
	}
}

func GetComplianceModule() *Module {
	return &Module{
		Name:    "compliance",
		Version: "??? (not implemented)",
		Env:     getInsightsCoreEnv(),
		Exec:    []string{"python3", "-m", "insights.client.phase.v2"},
		Commands: [][]string{
			{"compliance"},
			{"compliance", "collect"},
		},
		CollectCommand: []string{"compliance", "collect"},
	}
}

func GetMalwareModule() *Module {
	return &Module{
		Name:    "malware",
		Version: "??? (not implemented)",
		Env:     getInsightsCoreEnv(),
		Exec:    []string{"python3", "-m", "insights.client.phase.v2"},
		Commands: [][]string{
			{"malware"},
			{"malware", "collect"},
		},
		CollectCommand: []string{"malware", "collect"},
	}
}

// getInsightsCoreEnv sets up the environment for the Python subshell.
//
// It sets LC_ALL=C.UTF-8 to ensure we don't need to deal with non-supported locales.
// It sets PYTHONPATH to include the path to the Core Egg from its default location of
// InsightsCorePath, or to file defined by the EGG environment variable.
func getInsightsCoreEnv() []string {
	env := []string{"LC_ALL=C.UTF-8"}
	pythonPath := os.Getenv("PYTHONPATH")
	if egg := os.Getenv("EGG"); egg != "" {
		env = append(env, fmt.Sprintf("PYTHONPATH=%s", egg+":"+pythonPath))
	} else {
		env = append(env, fmt.Sprintf("PYTHONPATH=%s", InsightsCorePath+":"+pythonPath))
	}
	return env
}
