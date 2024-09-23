package modules

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func GetAdvisorModule() *Module {
	return &Module{
		Name:    "advisor",
		Version: getInsightsCoreVersion(),
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
		ArchiveCommand:     []string{"advisor", "collect"},
		ArchiveContentType: "application/vnd.redhat.advisor.collection",
	}
}

func GetComplianceModule() *Module {
	return &Module{
		Name:    "compliance",
		Version: getInsightsCoreVersion(),
		Env:     getInsightsCoreEnv(),
		Exec:    []string{"python3", "-m", "insights.client.phase.v2"},
		Commands: [][]string{
			{"compliance"},
			{"compliance", "collect"},
		},
		ArchiveCommand:     []string{"compliance", "collect"},
		ArchiveContentType: "application/vnd.redhat.compliance.something",
	}
}

func GetMalwareModule() *Module {
	return &Module{
		Name:    "malware",
		Version: getInsightsCoreVersion(),
		Env:     getInsightsCoreEnv(),
		Exec:    []string{"python3", "-m", "insights.client.phase.v2"},
		Commands: [][]string{
			{"malware"},
			{"malware", "collect"},
		},
		ArchiveCommand:     []string{"malware", "collect"},
		ArchiveContentType: "application/vnd.redhat.malware-detection.results",
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

var insightsCoreVersionIsCached = false
var insightsCoreVersion = ""

// getInsightsCoreVersion runs the Core to figure out what version it has.
//
// Since the Core does not have its metadata in a file, we need to read this dynamically.
func getInsightsCoreVersion() string {
	if insightsCoreVersionIsCached {
		return insightsCoreVersion
	}

	cmd := exec.Command("python3", "-c", "from insights.client import InsightsClient; print(InsightsClient(None, False).version())")
	for _, variable := range os.Environ() {
		cmd.Env = append(cmd.Env, variable)
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf(`PYTHONPATH=/etc/insights-client/rpm.egg:%s"`, os.Getenv(`PYTHONPATH`)))

	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	err := cmd.Run()

	if err != nil {
		slog.Error(
			"could not request Core version",
			slog.String("error", err.Error()),
			slog.Any("stdout", stdoutBuffer.String()),
			slog.Any("stderr", stderrBuffer.String()),
		)
		insightsCoreVersion = "??? (parsing error)"
		insightsCoreVersionIsCached = true
		return insightsCoreVersion
	}

	insightsCoreVersion = strings.TrimSpace(stdoutBuffer.String())
	insightsCoreVersionIsCached = true
	return insightsCoreVersion
}
