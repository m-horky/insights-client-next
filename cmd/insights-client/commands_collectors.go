package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	_ "github.com/briandowns/spinner"

	"github.com/m-horky/insights-client-next/public/collectors"
	"github.com/m-horky/insights-client-next/public/insights/services/ingress"
)

func runCollectorList() error {
	fmt.Println("Available collectors:")
	for _, collector := range collectors.GetCollectors() {
		fmt.Printf("* %s %s\n", collector.Name, collector.Version)
	}
	return nil
}

func runCollector(arguments Arguments) error {
	collector, err := collectors.GetCollector(arguments.Collector)
	if err != nil {
		return err
	}

	// TODO Check that we're registered first

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(collector.Exec, collector.ExecArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = collector.Env

	slog.Debug(
		"running a collector",
		slog.String("name", collector.Name),
		slog.String("version", collector.Version),
		slog.String("exec", fmt.Sprintf("%s %s", collector.Exec, strings.Join(collector.ExecArgs, " "))),
		slog.String("environment", strings.Join(collector.Env, " ")),
	)

	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	if isRichOutput(arguments) {
		spin.Suffix = fmt.Sprintf(" waiting for '%s' to collect its data", collector.Name)
		spin.Start()
	}
	err = cmd.Run()
	if isRichOutput(arguments) {
		spin.Stop()
	}

	if err != nil {
		return err
	}
	path := strings.TrimSpace(stdout.String())
	defer os.Remove(path)

	archive := ingress.Archive{ContentType: collector.ContentType, Path: path}

	if isRichOutput(arguments) {
		spin.Suffix = " uploading archive"
		spin.Start()
	}
	_, err = ingress.UploadArchive(archive)
	if isRichOutput(arguments) {
		spin.Stop()
	}

	if err != nil {
		return err
	}

	fmt.Println("Archive uploaded.")
	return nil
}
