package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/m-horky/insights-client-next/internal/app"
	"github.com/m-horky/insights-client-next/public/collectors"
)

func init() {
	initLogging()
	initCLI()
}

func initLogging() {
	fp, err := os.OpenFile(app.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	slog.SetDefault(slog.New(app.NewFileHandler(
		fp, &slog.HandlerOptions{AddSource: true, Level: app.GetConfiguration().LogLevel},
	)))
}

func initCLI() {
	cli.HelpFlag = &cli.BoolFlag{Name: "help"}
	cli.VersionFlag = &cli.BoolFlag{Name: "version"}
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Println(os.Args[0], app.Version)
	}
}

func main() {
	cmd := &cli.Command{
		Name:            os.Args[0],
		HideHelpCommand: true,
		Version:         app.Version,
		Usage:           "Upload data to Red Hat Insights",
		UsageText:       fmt.Sprintf("%s [COMMAND]", os.Args[0]),
		Action:          runCLI,
	}

	slog.Debug("started", slog.Any("args", os.Args))
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("finished", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Debug("finished")
}

func runCLI(_ context.Context, _ *cli.Command) error {
	if os.Geteuid() != 0 {
		fmt.Println("Error: This command has to be run with superuser privileges.")
		return nil
	}

	if _, err := os.Stat("/etc/insights-client/.registered"); err != nil {
		slog.Error(".registered file does not exist", slog.String("error", err.Error()))
		return nil
	}
	if _, err := os.Stat("/etc/insights-client/machine-id"); err != nil {
		slog.Error("machine-id file does not exist", slog.String("error", err.Error()))
		return nil
	}

	collector := collectors.GetDefaultCollector()
	fmt.Printf("Running collector '%s'.\n", collector.Name)
	slog.Debug("Running collector", slog.String("name", collector.Name))
	return collector.Run()
}
