package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/m-horky/insights-client-next/api"
	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/api/inventory"
	"github.com/m-horky/insights-client-next/app"
	"github.com/m-horky/insights-client-next/collectors"
	"github.com/m-horky/insights-client-next/internal"
)

func init() {
	initLogging()
	initCLI()
	initServices()
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
		slog.SetDefault(slog.New(internal.NewColorHandler(
			os.Stderr, &slog.HandlerOptions{AddSource: true, Level: internal.GetConfiguration().LogLevel},
		)))
	} else {
		fp, err := os.OpenFile(internal.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// We're not root, we can't log.
			// It doesn't matter the only commands that can be run as root are --version and --help.
			return
		}
		slog.SetDefault(slog.New(internal.NewFileHandler(
			fp, &slog.HandlerOptions{AddSource: true, Level: internal.GetConfiguration().LogLevel},
		)))
	}
}

func initCLI() {
	cli.HelpFlag = &cli.BoolFlag{Name: "help"}
	cli.VersionFlag = &cli.BoolFlag{Name: "version"}
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Println("insights-client", internal.Version)
		for _, collector := range collectors.GetCollectors() {
			fmt.Printf("* %s %s\n", collector.Name, collector.Version)
		}
	}
}

func initServices() {
	config := internal.GetConfiguration()
	{
		s := api.NewService(config.APIProtocol, config.APIHost, config.APIPort, "api/inventory/v1")
		s.Authenticate(config.IdentityCertificate, config.IdentityKey)
		inventory.Init(&s)
	}
	{
		s := api.NewService(config.APIProtocol, config.APIHost, config.APIPort, "api/ingress/v1")
		s.Authenticate(config.IdentityCertificate, config.IdentityKey)
		ingress.Init(&s)
	}
}

func main() {
	cmd := buildCLI()

	slog.Debug("started", slog.Any("args", os.Args))
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		if humanError, isHuman := err.(app.HumanError); isHuman {
			fmt.Println(humanError.Human())
		} else {
			fmt.Println("Error: " + err.Error())
		}
		slog.Error("finished", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Debug("finished")
}

func buildCLI() *cli.Command {
	return &cli.Command{
		Name:            "insights-client",
		HideHelpCommand: true,
		Version:         internal.Version,
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

func verifyCollector(_ context.Context, _ *cli.Command, collector string) error {
	if _, err := collectors.GetCollector(collector); err != nil {
		fmt.Printf("Error: invalid collector: '%s'\n", collector)
		return err
	}
	return nil
}

func verifyFormat(_ context.Context, _ *cli.Command, format string) error {
	if _, err := internal.ParseFormat(format); err != nil {
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
	Debug            bool
	Format           internal.Format
}

// parseCLI converts the cli.Command object into a clean structure.
func parseCLI(cmd *cli.Command) (*Arguments, error) {
	arguments := &Arguments{}

	// flags
	arguments.Format = internal.MustParseFormat(cmd.String("format"))
	arguments.Debug = cmd.IsSet("debug")

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

func runCLI(_ context.Context, cmd *cli.Command) error {
	arguments, err := parseCLI(cmd)
	if err != nil {
		return app.NewError(app.ErrInput, err, "Could not parse the input.")
	}

	if arguments.Help {
		_ = cli.ShowAppHelp(cmd)
		return nil
	}

	// ask for elevated privileges
	if os.Geteuid() != 0 {
		return app.NewError(app.ErrPermissions, nil, "This command has to be run with superuser privileges.")
	}

	// handle commands
	if arguments.Register {
		return runRegister(arguments)
	}
	if arguments.Unregister {
		return runUnregister()
	}
	if arguments.Status {
		return runStatus()
	}
	if arguments.DisplayName != "" || arguments.ResetDisplayName {
		return runDisplayName(arguments)
	}
	if arguments.AnsibleHost != "" || arguments.ResetAnsibleHost {
		return runAnsibleHostname(arguments)
	}
	if arguments.CollectorList {
		return runCollectorList()
	}
	if arguments.Collector != "" {
		return runCollector(*arguments)
	}

	return app.NewError(nil, nil, "Not implemented.")
}
