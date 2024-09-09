package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

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
	cmd.CustomRootCommandHelpTemplate = buildHelpText()

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

type commandCategory struct {
	Name     string
	Commands []cli.Flag
}

var commands = []commandCategory{
	{
		Name: "HOST",
		Commands: []cli.Flag{
			&cli.BoolFlag{Name: "register", Usage: "register the host"},
			&cli.BoolFlag{Name: "unregister", Usage: "unregister the host"},
			&cli.BoolFlag{Name: "status", Usage: "display host status"},
		},
	},
	{
		Name: "INVENTORY",
		Commands: []cli.Flag{
			&cli.BoolFlag{Name: "checkin", Usage: "send lightweight check-in notification"},
			&cli.StringFlag{Name: "display-name", Usage: "set display name of a host"},
			&cli.StringFlag{Name: "ansible-host", Usage: "set Ansible display name of a host"},
			&cli.StringFlag{Name: "group", Usage: "add system to Inventory group"},
		},
	},
	{
		Name: "DATA COLLECTION",
		Commands: []cli.Flag{
			&cli.StringFlag{Name: "collector", Usage: "run collector", Action: validateCollector},
			&cli.BoolFlag{Name: "collector-list", Aliases: []string{"list-collectors"}, Usage: "list collectors"},
			&cli.StringFlag{Name: "content-type", Usage: "content type for manual upload"},
			&cli.StringFlag{Name: "payload", Usage: "archive path for manual upload"},
			&cli.StringSliceFlag{Name: "collector-option", Aliases: []string{"opt"}, Usage: "collector option"},
			&cli.StringFlag{Name: "output-dir", Usage: "do not upload, collect into directory"},
			&cli.StringFlag{Name: "output-file", Usage: "do not upload, collect into file"},
		},
	},
	{
		Name: "GLOBAL FLAGS",
		Commands: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "human", Action: validateFormat, Usage: "change output format"},
			&cli.BoolFlag{Name: "debug", Usage: "print logs to stderr instead of a log file"},
		},
	},
	{
		Name: "DEPRECATED",
		Commands: []cli.Flag{
			&cli.UintFlag{Name: "retry", Usage: "ignored"},
			&cli.BoolFlag{Name: "validate", Usage: "ignored"},
			&cli.BoolFlag{Name: "quiet", Usage: "ignored"},
			&cli.BoolFlag{Name: "silent", Usage: "ignored"},
			&cli.StringFlag{Name: "conf", Aliases: []string{"c"}, Usage: "ignored"},
			&cli.StringFlag{Name: "compressor", Usage: "ignored"},
			&cli.BoolFlag{Name: "offline", Usage: "ignored"},
			&cli.StringFlag{Name: "logging-file", Usage: "ignored"},
			&cli.StringFlag{Name: "diagnosis", Usage: "alias for '--collector advisor --opt=diagnosis'"},
			&cli.BoolFlag{Name: "check-results", Usage: "alias for '--collector advisor --opt=check-results'"},
			&cli.BoolFlag{Name: "show-results", Usage: "alias for '--collector advisor --opt=show-results'"},
			&cli.BoolFlag{Name: "list-specs", Usage: "alias for '--collector advisor --opt=list-specs'"},
			&cli.BoolFlag{Name: "compliance", Usage: "alias for '--collector compliance'"},
			&cli.BoolFlag{Name: "test-connection", Usage: "alias for '--status'"},
			&cli.BoolFlag{Name: "no-upload", Usage: fmt.Sprintf("alias for '--output-file %sarchive-`date +%%s`'", collectors.ArchiveDirectory)},
			&cli.BoolFlag{Name: "keep-archive", Usage: fmt.Sprintf("alias for '--output-file %sarchive-`date +%%s`'", collectors.ArchiveDirectory)},
			&cli.BoolFlag{Name: "support", Usage: "alias for 'sosreport'"},
			&cli.BoolFlag{Name: "enable-schedule", Usage: "alias for '--register'"},
			&cli.BoolFlag{Name: "disable-schedule", Usage: "alias for '--unregister'"},
		},
	},
}

func buildHelpText() string {
	help := []string{`{{.Name}}, version {{.Version}}`}

	maxFlagLength := 0
	for _, cmdGrp := range commands {
		for _, cmd := range cmdGrp.Commands {
			if len(cmd.Names()[0]) > maxFlagLength {
				maxFlagLength = len(cmd.Names()[0])
			}
		}
	}

	for _, cmdGrp := range commands {
		help = append(help, ``)
		help = append(help, fmt.Sprintf("Category: %s", cmdGrp.Name))

		for _, cmd := range cmdGrp.Commands {
			// left-justify flag names
			help = append(help, fmt.Sprintf(
				"  --%s%s  %s",
				cmd.Names()[0],
				strings.Repeat(" ", maxFlagLength-len(cmd.Names()[0])),
				cmd.(cli.DocGenerationFlag).GetUsage(),
			))
		}
	}

	return strings.Join(help, "\n") + "\n"
}

func buildCLI() *cli.Command {
	var flags []cli.Flag
	for _, commandGroup := range commands {
		for _, cmd := range commandGroup.Commands {
			flags = append(flags, cmd)
		}
	}

	return &cli.Command{
		Name:            "insights-client",
		HideHelpCommand: true,
		Version:         internal.Version,
		Usage:           "Upload data to Red Hat Insights",
		UsageText:       fmt.Sprintf("%s COMMAND [FLAGS...]", "insights-client"),
		Flags:           flags,
		Action:          runCLI,
	}
}

func validateCollector(_ context.Context, _ *cli.Command, collector string) error {
	if _, err := collectors.GetCollector(collector); err != nil {
		fmt.Printf("Error: invalid collector: '%s'\n", collector)
		return err
	}
	return nil
}

func validateFormat(_ context.Context, _ *cli.Command, format string) error {
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
	Checkin          bool
	DisplayName      string
	ResetDisplayName bool
	AnsibleHost      string
	ResetAnsibleHost bool
	Group            string
	Collector        string
	CollectorOptions []string
	CollectorList    bool
	Payload          string
	ContentType      string
	OutputDir        string
	OutputFile       string
	Help             bool
	Debug            bool
	Format           internal.Format
}

// parseCLI converts the cli.Command object into a clean structure.
func parseCLI(cmd *cli.Command) (*Arguments, app.HumanError) {
	arguments := &Arguments{}

	// flags
	arguments.Format = internal.MustParseFormat(cmd.String("format"))
	arguments.Debug = cmd.IsSet("debug")
	arguments.Group = cmd.String("group")
	arguments.CollectorOptions = cmd.StringSlice("collector-option")
	arguments.OutputDir = cmd.String("output-dir")
	arguments.OutputFile = cmd.String("output-file")

	// display deprecation notices
	for oldCmd, newCmd := range map[string]string{"test-connection": "status", "compliance": "--collector compliance"} {
		if cmd.IsSet(oldCmd) {
			fmt.Printf("Notice: Command '%s' is deprecated, use '%s' instead.\n", oldCmd, newCmd)
		}
	}
	for _, ignored := range []string{"retry", "validate", "quiet", "silent", "conf", "compressor", "offline", "logging-file"} {
		if cmd.IsSet(ignored) {
			fmt.Printf("Notice: Command '%s' is deprecated and has no effect.\n", ignored)
		}
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
	if cmd.Bool("checkin") {
		arguments.Checkin = true
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
	if (cmd.IsSet("payload") && !cmd.IsSet("content-type")) || (!cmd.IsSet("payload") && cmd.IsSet("content-type")) {
		return nil, app.NewError(nil, nil, "Both --payload and --content-type have to be set.")
	}
	if cmd.IsSet("payload") && cmd.IsSet("payload") {
		arguments.Payload = cmd.String("payload")
		arguments.ContentType = cmd.String("content-type")
		return arguments, nil
	}

	slog.Debug("no command supplied, assuming 'help'")
	arguments.Help = true
	return arguments, nil
}

func runCLI(_ context.Context, cmd *cli.Command) error {
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
