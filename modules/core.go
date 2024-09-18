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
		Name:        "advisor",
		Version:     getCoreVersion(),
		Env:         getCoreEnvironment(),
		Exec:        "python3",
		ExecArgs:    []string{"-m", "insights.client.phase.v2", "advisor"},
		ContentType: "application/vnd.redhat.advisor.collection+tgz",
	}
}

// getCoreVersion
func getCoreVersion() string {
	if os.Geteuid() != 0 {
		return "??? (superuser permission required)"
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
		return "??? (parsing error)"
	}
	return strings.TrimSpace(stdoutBuffer.String())
}

// getCoreEnvironment sets up the environment for a Python subshell.
//
// It sets LC_ALL to C.UTF-8 to ensure we don't need to deal with non-supported locales.
// It sets PYTHONPATH to include the path to the egg from EGG environment
// variable (if set) or from the RPM egg (otherwise).
func getCoreEnvironment() []string {
	env := []string{"LC_ALL=C.UTF-8"}
	pythonPath := os.Getenv("PYTHONPATH")
	if egg := os.Getenv("EGG"); egg != "" {
		env = append(env, fmt.Sprintf("PYTHONPATH=%s", egg+":"+pythonPath))
	} else {
		env = append(env, fmt.Sprintf("PYTHONPATH=%s", "/etc/insights-client/rpm.egg:"+pythonPath))
	}
	return env
}
