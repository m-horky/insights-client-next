package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/m-horky/insights-client-next/internal/api"
	"github.com/m-horky/insights-client-next/internal/app"
	"github.com/m-horky/insights-client-next/public/collectors"
	"github.com/m-horky/insights-client-next/public/insights/api/inventory"
)

func init() {
	initLogging()
	initCli()
}

func initLogging() {
	debug := false
	for _, arg := range os.Args {
		if arg == "--debug" {
			debug = true
			break
		}
	}

	if debug {
		slog.SetDefault(slog.New(app.NewColorHandler(
			os.Stderr, &slog.HandlerOptions{AddSource: true, Level: app.GetConfiguration().LogLevel},
		)))
	} else {
		fp, err := os.OpenFile(app.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err.Error())
		}
		slog.SetDefault(slog.New(app.NewFileHandler(
			fp, &slog.HandlerOptions{AddSource: true, Level: app.GetConfiguration().LogLevel},
		)))
	}
}

func initCli() {
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

	slog.Debug("started", slog.Any("args", os.Args))
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
			&cli.StringFlag{Name: "format", Category: "global flags", Value: "human", Action: verifyFormat, Usage: "change output format"},
			&cli.BoolFlag{Name: "debug", Category: "global flags", Usage: "print logs to stderr instead of a log file"},
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
	Register         bool
	Unregister       bool
	Status           bool
	DisplayName      string
	ResetDisplayName bool
	AnsibleHost      string
	ResetAnsibleHost bool
	Collector        string
	CollectorList    bool
	Help             bool
	Format           app.Format
}

// parseCLI converts the cli.Command object into a clean structure.
func parseCLI(cmd *cli.Command) (*Arguments, error) {
	arguments := &Arguments{}

	// flags
	arguments.Format = app.MustParseFormat(cmd.String("format"))

	// display deprecation notices
	if cmd.Bool("test-connection") {
		fmt.Println("Warning: command 'test-connection' is deprecated and will be removed in future releases. Use 'status' instead.")
	}
	if cmd.Bool("compliance") {
		fmt.Println("Warning: command 'compliance' is deprecated and will be removed in future releases. Use '--collector compliance' instead.")
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
	if cmd.IsSet("display-name") {
		arguments.ResetDisplayName = true
		return arguments, nil
	}
	if cmd.String("ansible-host") != "" {
		arguments.AnsibleHost = cmd.String("ansible-host")
		return arguments, nil
	}
	if cmd.IsSet("ansible-host") {
		arguments.ResetAnsibleHost = true
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

	// ask for elevated privileges
	if os.Geteuid() != 0 {
		fmt.Println("Error: This command has to be run with superuser privileges.")
		return fmt.Errorf("this command has to be run with superuser privileges")
	}

	// handle commands
	if arguments.DisplayName != "" || arguments.ResetDisplayName {
		host, err := api.GetCurrentInventoryHost()
		if err != nil {
			fmt.Println("Error: could not get current inventory host.")
			return err
		}

		displayName := arguments.DisplayName
		if arguments.ResetDisplayName {
			displayName, err = os.Hostname()
			if err != nil {
				fmt.Printf("Error: Could not reset display name.")
				slog.Error("could not determine hostname", slog.String("error", err.Error()))
				return err
			}
		}

		err = inventory.UpdateDisplayName(host.InsightsInventoryID, displayName)
		if err != nil {
			fmt.Println("Error: Could not update display name.")
			return err
		}
		if arguments.ResetDisplayName {
			fmt.Println("Notice: Display name reset.")
		} else {
			fmt.Println("Notice: Display name updated.")
		}
		return nil
	}

	if arguments.AnsibleHost != "" || arguments.ResetAnsibleHost {
		host, err := api.GetCurrentInventoryHost()
		if err != nil {
			fmt.Println("Error: Could not get current Inventory host.")
			return err
		}

		ansibleHostname := arguments.AnsibleHost
		if arguments.ResetAnsibleHost {
			ansibleHostname, err = os.Hostname()
			if err != nil {
				fmt.Printf("Error: Could not reset Ansible hostname.")
				slog.Error("could not determine hostname", slog.String("error", err.Error()))
				return err
			}
		}

		err = inventory.UpdateAnsibleHostname(host.InsightsInventoryID, ansibleHostname)
		if err != nil {
			fmt.Println("Error: Could not update Ansible hostname.")
			return err
		}
		if arguments.ResetAnsibleHost {
			fmt.Println("Notice: Ansible hostname reset.")
		} else {
			fmt.Println("Notice: Ansible hostname updated.")
		}
		return nil
	}

	fmt.Println("Error: Not implemented.")
	return fmt.Errorf("not implemented")
}
