package collectors

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func GetAdvisorCollector() *Collector {
	return &Collector{
		Name:        "advisor",
		Version:     getCoreVersion(),
		Env:         []string{`PYTHONPATH=/etc/insights-client/rpm.egg`},
		Exec:        "python3",
		ExecArgs:    []string{"-c", "import sys; from insights.client.phase import v1 as client; phase = getattr(client, 'collect_and_output'); sys.exit(phase())"},
		ContentType: "application/vnd.redhat.advisor.collection+tgz",
	}
}

func getCoreVersion() string {
	if os.Geteuid() > 0 {
		slog.Warn("core requires root")
		return "unknown (no root)"
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
		return "unknown (parse error)"
	}
	return strings.TrimSpace(stdoutBuffer.String())
}
