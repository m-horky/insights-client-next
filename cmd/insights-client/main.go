package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/m-horky/insights-client-next/cmd/insights-client/actions"
	"github.com/m-horky/insights-client-next/internal/collectors"
	"github.com/m-horky/insights-client-next/internal/configuration"
	"github.com/m-horky/insights-client-next/internal/constants"
	"github.com/m-horky/insights-client-next/internal/enums"
	"github.com/m-horky/insights-client-next/internal/system"
)

func init() {
	config := configuration.GetConfiguration()
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stderr, &slog.HandlerOptions{Level: config.LogLevel},
	)))

	cli.HelpFlag = &cli.BoolFlag{Name: "help"}

	cli.VersionFlag = &cli.BoolFlag{Name: "version"}
	cli.VersionPrinter = func(cmd *cli.Command) {
		_ = actions.PrintVersion()
	}
}

func main() {
	cmd := &cli.Command{
		Name:            "insights-client",
		HideHelpCommand: true,
		Version:         constants.Version,
		Usage:           "Upload data to Red Hat Insights",
		UsageText:       fmt.Sprintf("%s COMMAND [FLAGS...]", os.Args[0]),
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "register", Category: "host", Usage: "register the host"},
			&cli.BoolFlag{Name: "unregister", Category: "host", Usage: "unregister the host"},
			&cli.BoolFlag{Name: "status", Category: "host", Usage: "display host status"},
			&cli.BoolFlag{Name: "display-name", Category: "host", Usage: "set host display name in Inventory"},
			&cli.BoolFlag{Name: "ansible-host", Category: "host", Usage: "set Ansible hostname in Inventory"},
			&cli.StringFlag{
				Name:     "collector",
				Aliases:  []string{"m"},
				Category: "data collection",
				Usage:    "run collector",
				Action: func(ctx context.Context, command *cli.Command, collector string) error {
					if _, err := collectors.GetCollector(collector); err != nil {
						fmt.Printf("Error: invalid collector: '%s'\n", collector)
						return err
					}
					return nil
				},
			},
			&cli.BoolFlag{Name: "collector-list", Category: "data collection", Usage: "list data collectors"},
			// Deprecated commands
			&cli.BoolFlag{Name: "test-connection", Category: "deprecated", Usage: "alias for '--status'"},
			&cli.BoolFlag{Name: "compliance", Category: "deprecated", Usage: "alias for '--collector compliance'"},
			// Flags
			&cli.StringFlag{
				Name:     "format",
				Category: "global flags",
				Value:    "human",
				Action: func(ctx context.Context, cmd *cli.Command, format string) error {
					if format != "human" && format != "json" {
						fmt.Printf("Error: invalid format: '%s'\n", format)
						return fmt.Errorf("invalid format: %s", format)
					}
					return nil
				},
			},
		},
		Action: run,
	}

	slog.Debug("program started", slog.Any("args", os.Args))
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("program crashed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Debug("program finished")
}

// run acts as an action router.
func run(ctx context.Context, cmd *cli.Command) error {
	// Flags
	if _, err := enums.ParseFormat(cmd.String("format")); err != nil {
		slog.Error("could not parse format", slog.Any("error", err))
		return err
	}

	// Deprecated commands
	for _, flag := range []string{"test-connection"} {
		if cmd.IsSet(flag) {
			slog.Warn("flag is deprecated", slog.String("flag", flag))
		}
	}

	// Commands
	if cmd.Bool("register") {
		if _, err := system.GetInventoryHost(); err == nil {
			fmt.Printf("Error: This host is already registered.\n")
			return nil
		}
		slog.Warn("register: not implemented")
		return nil
	}
	if cmd.Bool("unregister") {
		slog.Warn("unregister: not implemented")
		return nil
	}
	if cmd.IsSet("display-name") {
		return actions.SetDisplayName(cmd.String("display-name"))
	}
	if cmd.Bool("status") || cmd.Bool("test-connection") {
		slog.Warn("status: not implemented")
		return nil
	}

	if cmd.IsSet("collector") {
		return actions.RunCollector(cmd.String("collector"))
	}
	if cmd.IsSet("collector-list") {
		return actions.ListCollectors()
	}

	slog.Debug("no argument specified, running default collector")
	return actions.RunCollector(constants.DefaultCollector)
}
