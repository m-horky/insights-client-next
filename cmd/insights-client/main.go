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
	// TODO Write to log file instead.
	// TODO Include file in the log statement.
	// TODO For --verbose, pretty-print to stderr?
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stderr, &slog.HandlerOptions{Level: app.GetConfiguration().LogLevel},
	)))

	cli.HelpFlag = &cli.BoolFlag{Name: "help"}
	cli.VersionFlag = &cli.BoolFlag{Name: "version"}
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Println("insights-client", app.Version)

		for _, collector := range collectors.GetCollectors() {
			fmt.Printf("* %s %s\n", collector.Name, collector.Version)
		}
	}
}

func main() {
	cmd := buildCLI()

	slog.Debug("program started", slog.Any("args", os.Args))
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("finished", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Debug("finished")
}

func buildCLI() *cli.Command {
	return &cli.Command{
		Name:            "insights-client",
		HideHelpCommand: true,
		Version:         app.Version,
		Usage:           "Upload data to Red Hat Insights",
		UsageText:       fmt.Sprintf("%s COMMAND [FLAGS...]", "insights-client"),
		Flags: []cli.Flag{
			// client
			&cli.BoolFlag{Name: "register", Category: "host", Usage: "register the host"},
			&cli.BoolFlag{Name: "unregister", Category: "host", Usage: "unregister the host"},
			&cli.BoolFlag{Name: "status", Category: "host", Usage: "display host status"},

			// inventory
			&cli.StringFlag{Name: "display-name", Category: "inventory", Usage: "set display name of a host"},
			&cli.StringFlag{Name: "ansible-host", Category: "inventory", Usage: "set Ansible display name of a host"},

			// data collection
			&cli.StringFlag{Name: "collector", Aliases: []string{"m"}, Category: "data collection", Usage: "run collector", Action: verifyCollector},
			&cli.BoolFlag{Name: "collector-list", Aliases: []string{"list-collectors"}, Category: "data collection", Usage: "list collectors"},

			// deprecated
			&cli.BoolFlag{Name: "test-connection", Category: "deprecated", Usage: "alias for '--status'"},
			&cli.BoolFlag{Name: "compliance", Category: "deprecated", Usage: "alias for '--collector compliance'"},

			// flags
			&cli.StringFlag{Name: "format", Category: "global flags", Value: "human", Action: verifyFormat},
		},
		Action: runCLI,
	}
}

func verifyCollector(ctx context.Context, cmd *cli.Command, collector string) error {
	if _, err := collectors.GetCollector(collector); err != nil {
		fmt.Printf("Error: invalid collector: '%s'\n", collector)
		return err
	}
	return nil
}

func verifyFormat(_ context.Context, _ *cli.Command, format string) error {
	if _, err := app.ParseFormat(format); err != nil {
		fmt.Printf("Error: invalid format: '%s'\n", format)
		return fmt.Errorf("invalid format: '%s'", format)
	}
	return nil
}

type Arguments struct {
	Register      bool
	Unregister    bool
	Status        bool
	DisplayName   string
	AnsibleHost   string
	Collector     string
	CollectorList bool
	Help          bool
	Format        app.Format
}

// parseCLI converts the cli.Command object into a clean structure.
func parseCLI(cmd *cli.Command) (*Arguments, error) {
	arguments := &Arguments{}

	// flags
	arguments.Format = app.MustParseFormat(cmd.String("format"))

	// display deprecation notices
	if cmd.Bool("test-connection") {
		fmt.Println("Notice: command 'test-connection' is deprecated and will be removed in future releases. Use 'status' instead.")
	}
	if cmd.Bool("compliance") {
		fmt.Println("Notice: command 'compliance' is deprecated and will be removed in future releases. Use '--collector compliance' instead.")
	}

	// client
	if cmd.Bool("register") {
		arguments.Register = true
		return arguments, nil
	}
	if cmd.Bool("unregister") {
		arguments.Unregister = true
		return arguments, nil
	}
	if cmd.Bool("status") || cmd.Bool("test-connection") {
		arguments.Status = true
		return arguments, nil
	}

	// inventory
	if cmd.String("display-name") != "" {
		arguments.DisplayName = cmd.String("display-name")
		return arguments, nil
	}
	if cmd.String("ansible-host") != "" {
		arguments.AnsibleHost = cmd.String("ansible-host")
		return arguments, nil
	}

	// data collection
	if cmd.Bool("collector-list") {
		arguments.CollectorList = true
		return arguments, nil
	}
	if cmd.String("collector") != "" {
		arguments.Collector = cmd.String("collector")
		return arguments, nil
	}
	if cmd.Bool("compliance") {
		arguments.Collector = "compliance"
		return arguments, nil
	}

	slog.Debug("no command supplied, assuming 'help'")
	arguments.Help = true
	return arguments, nil
}

func runCLI(ctx context.Context, cmd *cli.Command) error {
	arguments, err := parseCLI(cmd)
	if err != nil {
		return err
	}

	if arguments.Help {
		_ = cli.ShowAppHelp(cmd)
		return nil
	}

	fmt.Println("Error: not implemented")
	return fmt.Errorf("not implemented")
}
