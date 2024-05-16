package main

import (
	"fmt"
	"github.com/m-horky/insights-client-next/internal/api/inventory"
	"github.com/m-horky/insights-client-next/internal/system"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/m-horky/insights-client-next/internal/configuration"
	"github.com/m-horky/insights-client-next/internal/constants"
	"github.com/m-horky/insights-client-next/internal/core"
	"github.com/m-horky/insights-client-next/internal/enums"
)

func init() {
	config := configuration.GetConfiguration()
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stderr, &slog.HandlerOptions{Level: config.LogLevel},
	)))
}

func main() {
	app := cli.NewApp()
	app.HideHelpCommand = true
	app.HideHelp = true
	app.HideVersion = true
	app.Action = run
	app.CustomAppHelpTemplate = getHelp()

	app.Name = "insights-client"
	app.Usage = "Manage Red Hat Insights."
	app.Version = constants.Version

	app.Flags = []cli.Flag{
		// modifiers
		&cli.StringFlag{
			Name:  "format",
			Value: "human",
			Action: func(context *cli.Context, format string) error {
				if _, err := enums.ParseFormat(format); err != nil {
					return err
				}
				return nil
			},
		},
		// commands
		&cli.BoolFlag{Name: "register"},
		&cli.BoolFlag{Name: "unregister"},
		&cli.StringFlag{Name: "display-name"},
		&cli.StringFlag{Name: "ansible-host"},
		&cli.BoolFlag{Name: "version"},
		&cli.BoolFlag{Name: "status"},
		&cli.BoolFlag{Name: "help"},
		&cli.StringFlag{
			Name: "collector", Value: "advisor",
			Action: func(context *cli.Context, collector string) error {
				return core.VerifyCollector(collector)
			},
		},
		// TODO --collector-list?
		// deprecated commands
		&cli.BoolFlag{Name: "test-connection"},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("could not run insights-client", slog.Any("error", err))
	}
}

// getHelp creates custom help output
// because urfave/cli/v2 doesn't seem to be flexible enough to not support flag-only setup.
func getHelp() string {
	builder := strings.Builder{}

	builder.WriteString("Usage: insights-client COMMAND [FLAGS...]\n")
	builder.WriteString("\n")
	builder.WriteString("Manage data collection for Red Hat Insights.\n")

	// We can lie a bit; we can pretend --display-name is only available during registration.
	builder.WriteString("\nCOMMANDS\n")
	builder.WriteString("--register               register the system\n")
	builder.WriteString("  [--display-name NAME]    and set Insights hostname\n")
	builder.WriteString("  [--ansible-host NAME]    and set Ansible hostname\n")
	builder.WriteString("--unregister             unregister the host\n")
	builder.WriteString("--status                 display system status\n")
	builder.WriteString("--collector APP [...]    run collector\n")
	builder.WriteString("--version                display program version\n")
	builder.WriteString("--help                   display this help\n")

	builder.WriteString("\nDEPRECATED COMMANDS\n")
	builder.WriteString("--test-connection        use --status\n")
	builder.WriteString("--compliance [...]       use --collector compliance [...]\n")

	builder.WriteString("\nFLAGS\n")
	builder.WriteString("--format FORMAT          output format (options: human, json)\n")

	return builder.String()
}

// run acts as an action router.
func run(c *cli.Context) error {
	// Flags
	if _, err := enums.ParseFormat(c.String("format")); err != nil {
		slog.Error("could not parse format", slog.Any("error", err))
		return err
	}

	// Deprecated commands
	for _, flag := range []string{"test-connection"} {
		if c.IsSet(flag) {
			slog.Warn("flag is deprecated", slog.String("flag", flag))
		}
	}

	// Commands
	if c.Bool("version") {
		fmt.Printf("Insights Client: %s\n", constants.Version)
		fmt.Printf("Insights Core:   none\n")
		return nil
	}
	if c.Bool("register") {
		if _, err := system.GetInventoryHost(); err == nil {
			fmt.Printf("Error: This host is already registered.\n")
			return nil
		}
		slog.Warn("register: not implemented")
		return nil
	}
	if c.Bool("unregister") {
		slog.Warn("unregister: not implemented")
		return nil
	}
	if c.IsSet("display-name") {
		host, err := system.GetInventoryHost()
		if err != nil {
			fmt.Println("Error: Host is not registered.")
			return nil
		}
		err = inventory.UpdateDisplayName(host.InsightsInventoryID, c.String("display-name"))
		if err != nil {
			return err
		}
		fmt.Println("OK: Display name has been updated.")
	}
	if c.Bool("status") || c.Bool("test-connection") {
		slog.Warn("status: not implemented")
		return nil
	}
	if c.IsSet("collector") {
		slog.Warn("collector: not implemented")
		return nil
	}

	// FIXME Implicitly we should run Advisor collection instead.
	return nil
}
