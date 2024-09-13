package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	url := api.NewServiceURL(config.APIProtocol, config.APIHost, config.APIPort)
	inventory.Init(api.NewServiceWithAuthentication(url, config.IdentityCertificate, config.IdentityKey))
	ingress.Init(api.NewServiceWithAuthentication(url, config.IdentityCertificate, config.IdentityKey))
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
			&cli.BoolFlag{Name: "checkin", Usage: "send lightweight check-in notification"},
		},
	},
	{
		Name: "INVENTORY",
		Commands: []cli.Flag{
			&cli.StringFlag{Name: "display-name", Usage: "set display name of a host"},
			&cli.StringFlag{Name: "ansible-host", Usage: "set Ansible display name of a host"},
			&cli.StringFlag{Name: "group", Usage: "add system to Inventory group"},
		},
	},
	{
		Name: "DATA COLLECTION",
		Commands: []cli.Flag{
			&cli.StringFlag{Name: "collector", Aliases: []string{"m"}, Usage: "run collector and upload its archive", Action: validateCollector},
			&cli.BoolFlag{Name: "collector-list", Usage: "list collectors"},
			&cli.StringSliceFlag{Name: "collector-option", Aliases: []string{"opt"}, Usage: "set collector option"},
			&cli.StringFlag{Name: "output-dir", Usage: "do not upload, collect into directory"},
			&cli.StringFlag{Name: "output-file", Usage: "do not upload, collect into file"},
			&cli.StringFlag{Name: "payload", Usage: "upload archive from this path"},
			&cli.StringFlag{Name: "content-type", Usage: "upload archive with this content type"},
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
			&cli.StringFlag{Name: "diagnosis", Usage: "alias for '-m advisor --opt=diagnosis'"},
			&cli.BoolFlag{Name: "check-results", Usage: "alias for '-m advisor --opt=check-results'"},
			&cli.BoolFlag{Name: "show-results", Usage: "alias for '-m advisor --opt=show-results'"},
			&cli.BoolFlag{Name: "list-specs", Usage: "alias for '-m advisor --opt=list-specs'"},
			&cli.BoolFlag{Name: "compliance", Usage: "alias for '-m compliance'"},
			&cli.BoolFlag{Name: "test-connection", Usage: "alias for '--status'"},
			&cli.BoolFlag{Name: "no-upload", Usage: "alias for '--output-file [PATH]'"},
			&cli.BoolFlag{Name: "keep-archive", Usage: "alias for '--output-file [PATH]'"},
			&cli.BoolFlag{Name: "support", Usage: "alias for 'sosreport'"},
			&cli.BoolFlag{Name: "enable-schedule", Usage: "alias for '--register'"},
			&cli.BoolFlag{Name: "disable-schedule", Usage: "alias for '--unregister'"},
		},
	},
}

func buildHelpText() string {
	// FIXME Can we make this not break in narrow terminals?
	help := []string{`Usage: {{.Name}} [COMMAND] [FLAGS]`}

	maxFlagLength := 0
	for _, cmdGrp := range commands {
		for _, cmd := range cmdGrp.Commands {
			if len(buildHelpFlag(cmd)) > maxFlagLength {
				maxFlagLength = len(buildHelpFlag(cmd))
			}
		}
	}

	for _, cmdGrp := range commands {
		help = append(help, ``)
		help = append(help, fmt.Sprintf("Category: %s", cmdGrp.Name))

		for _, cmd := range cmdGrp.Commands {
			// left-justify flag names
			help = append(help, fmt.Sprintf(
				"  %s%s  %s",
				buildHelpFlag(cmd),
				strings.Repeat(" ", maxFlagLength-len(buildHelpFlag(cmd))),
				cmd.(cli.DocGenerationFlag).GetUsage(),
			))
		}
	}

	return strings.Join(help, "\n") + "\n"
}

// buildHelpFlag constructs a string out of the flag and its aliases
func buildHelpFlag(flag cli.Flag) string {
	result := "--" + flag.Names()[0]

	for _, alias := range flag.Names()[1:] {
		if len(alias) == 1 {
			result += ", -" + alias
		} else {
			result += ", --" + alias
		}
	}
	return result
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
	CheckIn          bool
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

// validateCLI performs input validation.
//
// It shows notices for commands that are deprecated.
//
// It ensures flags that assume other flags are properly joined.
func validateCLI(cmd *cli.Command) app.HumanError {
	// display deprecation notices
	for oldCmd, newCmd := range map[string]string{
		"diagnosis":        "--collector advisor --opt=diagnosis",
		"check-results":    "--collector advisor --opt=check-results",
		"show-results":     "--collector advisor --opt=show-results",
		"list-specs":       "--collector advisor --opt=list-specs",
		"compliance":       "--collector compliance",
		"test-connection":  "--status",
		"no-upload":        fmt.Sprintf("--output-file %sarchive-`date +%%s`", collectors.ArchiveDirectory),
		"keep-archive":     fmt.Sprintf("--output-file %sarchive-`date +%%s`", collectors.ArchiveDirectory),
		"support":          "sosreport",
		"enable-schedule":  "--register",
		"disable-schedule": "--unregister",
	} {
		if cmd.IsSet(oldCmd) {
			fmt.Printf("Notice: Flag '--%s' is deprecated, use '%s' instead.\n", oldCmd, newCmd)
		}
	}
	for _, ignored := range []string{"retry", "validate", "quiet", "silent", "conf", "compressor", "offline", "logging-file"} {
		if cmd.IsSet(ignored) {
			fmt.Printf("Notice: Flag '--%s' is deprecated and has no effect.\n", ignored)
		}
	}
	// validate input: first flag requires others
	for _, flags := range [][]string{
		{"content-type", "payload"},
		{"payload", "content-type"},
		{"output-dir", "collector"},
		{"output-file", "collector"},
	} {
		if !cmd.IsSet(flags[0]) {
			continue
		}
		for _, otherFlag := range flags[1:] {
			if !cmd.IsSet(otherFlag) {
				return app.NewError(app.ErrInput, nil, fmt.Sprintf(
					"Flag '--%s' also requires '--%s'.", flags[0], strings.Join(flags[1:], "', --'"),
				))
			}
		}
	}
	// FIXME This works but isn't nice, is there a better way?
	// validate input: can't be used together
	for _, flags := range [][]string{
		// top-level commands with other top-level commands
		{"register", "unregister", "status", "checkin", "collector", "collector-list"},
		// top-level commands with collector modifiers
		{"register", "unregister", "status", "checkin", "collector-option"},
		// top-level commands with upload modifiers
		{"register", "unregister", "status", "checkin", "output-dir", "output-file"},
		// 'group' can only be used alone or with 'register'
		{"group", "unregister", "status", "checkin", "collector", "collector-list", "payload"},
		// collection flags
		{"output-dir", "output-file"},
	} {
		var usedFlags []string
		for _, flag := range flags {
			if cmd.IsSet(flag) {
				usedFlags = append(usedFlags, flag)
			}
		}
		if len(usedFlags) > 1 {
			return app.NewError(app.ErrInput, nil, fmt.Sprintf(
				"Some flags can't be used together: '--%s'.", strings.Join(usedFlags, "', '--")),
			)
		}
	}

	return nil
}

// parseCLI converts the cli.Command object into a clean structure.
func parseCLI(cmd *cli.Command) (*Arguments, app.HumanError) {
	arguments := &Arguments{}

	// flags
	arguments.Format = internal.MustParseFormat(cmd.String("format"))
	arguments.Debug = cmd.IsSet("debug")
	arguments.OutputDir = cmd.String("output-dir")
	arguments.OutputFile = cmd.String("output-file")
	if cmd.IsSet("keep-archive") || cmd.IsSet("no-upload") {
		arguments.OutputFile = filepath.Join(collectors.ArchiveDirectory, fmt.Sprintf("archive-%d.tar.xz", time.Now().Unix()))
	}
	for _, option := range cmd.StringSlice("collector-option") {
		arguments.CollectorOptions = append(arguments.CollectorOptions, "--"+option)
	}

	// client
	if cmd.IsSet("register") {
		arguments.Register = true
		arguments.Collector = collectors.GetDefaultCollector().Name
		arguments.DisplayName = cmd.String("display-name")
		arguments.AnsibleHost = cmd.String("ansible-host")
		arguments.Group = cmd.String("group")
		return arguments, nil
	}
	if cmd.IsSet("unregister") {
		arguments.Unregister = true
		return arguments, nil
	}
	if cmd.IsSet("status") || cmd.IsSet("test-connection") {
		arguments.Status = true
		return arguments, nil
	}
	if cmd.IsSet("checkin") {
		arguments.CheckIn = true
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
	if cmd.IsSet("group") {
		arguments.Group = cmd.String("group")
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
	if cmd.IsSet("payload") && cmd.IsSet("payload") {
		arguments.Payload = cmd.String("payload")
		arguments.ContentType = cmd.String("content-type")
		return arguments, nil
	}

	slog.Debug("no command supplied, defaulting to data collection")
	arguments.Collector = collectors.GetDefaultCollector().Name
	return arguments, nil
}

func runCLI(_ context.Context, cmd *cli.Command) error {
	if err := validateCLI(cmd); err != nil {
		return err
	}
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
	if len(arguments.Group) > 0 {
		return runGroup(arguments)
	}
	if arguments.CollectorList {
		return runCollectorList()
	}
	if arguments.Collector != "" {
		return runCollector(arguments)
	}
	if arguments.Payload != "" && arguments.ContentType != "" {
		return runUploadExistingArchive(arguments)
	}

	return app.NewError(nil, nil, "Not implemented.")
}
