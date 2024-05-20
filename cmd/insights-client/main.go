package main

import (
	"context"
	"fmt"
	"github.com/m-horky/insights-client-next/internal/api/inventory"
	"github.com/m-horky/insights-client-next/internal/system"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"

	"github.com/m-horky/insights-client-next/internal/configuration"
	"github.com/m-horky/insights-client-next/internal/constants"
	"github.com/m-horky/insights-client-next/internal/enums"
)

func init() {
	config := configuration.GetConfiguration()
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stderr, &slog.HandlerOptions{Level: config.LogLevel},
	)))

	cli.HelpFlag = &cli.BoolFlag{Name: "help"}

	cli.VersionFlag = &cli.BoolFlag{Name: "version"}
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("Insights Client: %s\n", constants.Version)
		fmt.Printf("Insights Core:   none\n")
	}
}

func main() {
	cmd := &cli.Command{
		Name:            "insights-client",
		HideHelpCommand: true,
		Version:         constants.Version,
		Usage:           "Manage data connection for Red Hat Insights",
		UsageText:       fmt.Sprintf("%s COMMAND [FLAGS...]", os.Args[0]),
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "register", Category: "host", Usage: "register the host"},
			&cli.BoolFlag{Name: "unregister", Category: "host", Usage: "unregister the host"},
			&cli.BoolFlag{Name: "status", Category: "host", Usage: "display host status"},
			&cli.BoolFlag{Name: "display-name", Category: "host", Usage: "set host display name in Inventory"},
			&cli.BoolFlag{Name: "ansible-host", Category: "host", Usage: "set Ansible hostname in Inventory"},
			&cli.StringFlag{Name: "collector", Category: "data collection", Usage: "run collector"},
			// Deprecated commands
			&cli.BoolFlag{Name: "test-connection", Category: "deprecated", Usage: "alias for '--status'"},
			&cli.BoolFlag{Name: "compliance", Category: "deprecated", Usage: "alias for '--collector compliance'"},
			// Flags
			&cli.StringFlag{
				Name:     "format",
				Category: "global flags",
				Value:    "human",
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					if s != "human" && s != "json" {
						fmt.Printf("Error: invalid format: '%s'\n", s)
						return fmt.Errorf("invalid format: %s", s)
					}
					return nil
				},
			},
		},
		Action: run,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("crashed", slog.Any("error", err))
		os.Exit(1)
	}
}

// run acts as an action router.
func run(ctx context.Context, cmd *cli.Command) error {
	cli.DefaultAppComplete(ctx, cmd)

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
		host, err := system.GetInventoryHost()
		if err != nil {
			fmt.Println("Error: Host is not registered.")
			return nil
		}
		err = inventory.UpdateDisplayName(host.InsightsInventoryID, cmd.String("display-name"))
		if err != nil {
			return err
		}
		fmt.Println("OK: Display name has been updated.")
	}
	if cmd.Bool("status") || cmd.Bool("test-connection") {
		slog.Warn("status: not implemented")
		return nil
	}
	if cmd.IsSet("collector") {
		slog.Warn("collector: not implemented")
		return nil
	}

	// FIXME Implicitly we should run Advisor collection instead.
	slog.Warn("default: not implemented")
	return nil
}
