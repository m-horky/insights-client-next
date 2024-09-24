package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/m-horky/insights-client-next/api"
	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/api/inventory"
	"github.com/m-horky/insights-client-next/internal"
	"github.com/m-horky/insights-client-next/internal/impl"
	"github.com/m-horky/insights-client-next/modules"
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
		for _, module := range modules.GetModules() {
			fmt.Printf("* %s %s\n", module.Name, module.Version)
		}
	}
}

func initServices() {
	config := internal.GetConfiguration()
	// FIXME This won't work for IPv6 address
	address := &url.URL{Scheme: config.APIProtocol, Host: fmt.Sprintf("%s:%d", config.APIHost, config.APIPort)}
	// TODO Support the configuration file
	template := api.NewService(address).WithAuthentication(config.IdentityCertificate, config.IdentityKey).WithProxy(os.Getenv("HTTP_PROXY"))
	inventory.Init(template)
	ingress.Init(template)
}

func main() {
	cmd := buildCLI()
	cmd.CustomRootCommandHelpTemplate = buildHelpText()

	slog.Debug("started", slog.Any("args", os.Args))
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		if humanError, isHuman := err.(internal.IError); isHuman {
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
			&cli.StringSliceFlag{Name: "module", Aliases: []string{"m", "collector"}, Usage: "run module and upload its archive", Action: validateModule},
			&cli.BoolFlag{Name: "module-list", Usage: "list modules"},
			&cli.StringSliceFlag{Name: "module-option", Aliases: []string{"opt"}, Usage: "set module option"},
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
			&cli.BoolFlag{Name: "diagnosis", Usage: "alias for '-m advisor --opt=diagnosis'"},
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

// validateModule ensures that module exposes given subcommand.
func validateModule(_ context.Context, _ *cli.Command, command []string) error {
	for _, module := range modules.GetModules() {
		for _, cmd := range module.Commands {
			if reflect.DeepEqual(cmd, command) {
				return nil
			}
		}
	}

	fmt.Printf("Error: No module implements requested command '%s'.", strings.Join(command, " "))
	return errors.New("no module implements requested command")
}

func validateFormat(_ context.Context, _ *cli.Command, format string) error {
	if _, err := internal.ParseFormat(format); err != nil {
		fmt.Printf("Error: invalid format: '%s'\n", format)
		return fmt.Errorf("invalid format: '%s'", format)
	}
	return nil
}

// validateCLI performs input validation.
//
// It shows notices for flags that are deprecated.
//
// It ensures flags that assume other flags are properly joined.
func validateCLI(cmd *cli.Command) internal.IError {
	globalFlags := []string{"format", "debug"}

	// this includes the list of all valid combinations
	flagCombinations := [][]string{
		{"register"},
		{"register", "display-name"},
		{"register", "ansible-host"},
		{"register", "display-name", "ansible-host"},
		{"register", "group"},
		{"register", "group", "display-name"},
		{"register", "group", "ansible-host"},
		{"register", "group", "display-name", "ansible-host"},
		{"unregister"},
		{"status"},
		{"checkin"},
		{"display-name"},
		{"ansible-host"},
		{"group"},
		{"group", "offline"},
		{"module"},
		{"module", "module-option"},
		{"module-list"},
		{"output-dir"},
		{"output-file"},
		{"module", "output-dir"},
		{"module", "output-file"},
		{"module", "module-option", "output-dir"},
		{"module", "module-option", "output-file"},
		{"payload", "content-type"},
	}

	// key holds the primary flag we match by, the rest is modifiers
	setFlags := make(map[string]bool)

	for _, flag := range cmd.Flags {
		flagName := flag.Names()[0]

		if cmd.IsSet(flagName) {
			// resolve aliased commands
			resolvedFlagName, resolvedFlagNameOptions := resolveAlias(flagName)
			// announce deprecated & with no effect
			if resolvedFlagName == "" {
				fmt.Printf("Notice: Flag '--%s' is deprecated and has no effect.\n", flagName)
				continue
			}
			// announce deprecated & aliased
			if resolvedFlagName != flagName {
				if resolvedFlagNameOptions != "" {
					resolvedFlagName += "" + resolvedFlagNameOptions
				}
				fmt.Printf("Notice: Flag '--%s' is deprecated, use '--%s' instead.\n", flagName, resolvedFlagName)
			}

			// we don't need to check global flags, they can be applied to everything
			flagIsGlobal := false
			for _, globalFlag := range globalFlags {
				if resolvedFlagName == globalFlag {
					flagIsGlobal = true
					break
				}
			}
			if flagIsGlobal {
				continue
			}

			if _, found := setFlags[resolvedFlagName]; found {
				// resolved flags conflict (e.g. `--compliance --diagnosis`)
				return internal.NewError(internal.ErrInput, errors.New("found conflict in module flags"), "This flag combination is not valid.")
			}
			setFlags[resolvedFlagName] = true
		}
	}

	// exit immediately if we find combination match: validation is complete
	var finalFlags []string
	for flag := range setFlags {
		finalFlags = append(finalFlags, flag)
	}
	for _, combination := range flagCombinations {
		if reflect.DeepEqual(combination, finalFlags) {
			return nil
		}
	}

	// TODO Display all flag combinations that use the entered flags

	return internal.NewError(internal.ErrInput, errors.New("found generic flag conflict"), "This flag combination is not valid.")
}

// resolveAlias is used by validateCLI during flag combination check.
//
// A flag may be defined as an alias; this ensures we do not need to check them.
//
// Returns the input if the command is not an alias, empty string if the alias is noop, new string for resolved flag.
// Because this feeds directly into text notification about the alias, a slice is always returned. Only the first
// item will be used for comparison, full slice is used to display the help text.
func resolveAlias(flag string) (string, string) {
	switch flag {
	case "retry", "validate", "quiet", "silent", "conf", "compressor", "logging-file":
		return "", ""
	case "diagnosis", "check-results", "show-results", "list-specs":
		return "module", "=advisor --module-option=" + flag
	case "compliance":
		return "module", "=compliance"
	case "test-connection":
		return "status", ""
	case "no-upload", "keep-archive":
		return "output-file", ""
	case "enable-schedule":
		return "register", ""
	case "disable-schedule":
		return "unregister", ""
	}
	return flag, ""
}

// parseCLI converts the cli.Command object into a clean structure.
func parseCLI(cmd *cli.Command) *impl.Input {
	input := &impl.Input{}
	input.Format = internal.MustParseFormat(cmd.String("format"))
	input.Debug = cmd.IsSet("debug")

	if cmd.IsSet("help") {
		input.Action = impl.AHelp
	}

	// host
	if cmd.IsSet("register") && input.Action == impl.ANone {
		input.Action = impl.ARegister
		input.RegisterArgs = impl.ARegisterArgs{
			Group:           cmd.String("group"),
			DisplayName:     cmd.String("display-name"),
			AnsibleHostname: cmd.String("ansible-host"),
		}
	}
	if cmd.IsSet("unregister") && input.Action == impl.ANone {
		input.Action = impl.AUnregister
	}
	if cmd.IsSet("status") && input.Action == impl.ANone {
		input.Action = impl.AStatus
	}
	if cmd.IsSet("checkin") && input.Action == impl.ANone {
		input.Action = impl.ACheckIn
	}

	// inventory
	// TODO Enforce values for those, do not support reset, it should be explicit
	//  (unless the ansible role does something with this)
	if cmd.IsSet("display-name") && input.Action == impl.ANone {
		if cmd.String("display-name") == "" {
			input.Action = impl.AResetDisplayName
		} else {
			input.Action = impl.ASetDisplayName
			input.SetDisplayNameArgs = impl.ASetDisplayNameArgs{Name: cmd.String("display-name")}
		}
	}
	if cmd.IsSet("ansible-host") && input.Action == impl.ANone {
		if cmd.String("ansible-host") == "" {
			input.Action = impl.AResetAnsibleHostname
		} else {
			input.Action = impl.ASetAnsibleHostname
			input.SetAnsibleHostnameArgs = impl.ASetAnsibleHostnameArgs{Name: cmd.String("ansible-host")}
		}
	}

	// modules
	if cmd.IsSet("module-list") && input.Action == impl.ANone {
		input.Action = impl.AListModules
	}
	if (cmd.IsSet("module") || cmd.IsSet("collector")) && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		collector := cmd.StringSlice("collector")
		if cmd.IsSet("module") {
			collector = cmd.StringSlice("module")
		}
		input.RunModuleArgs = impl.ARunModuleArgs{Name: collector}
	}
	if cmd.IsSet("payload") && cmd.IsSet("content-type") && input.Action == impl.ANone {
		input.Action = impl.AUploadLocalArchive
		input.UploadLocalArchiveArgs = impl.AUploadLocalArchiveArgs{
			Path:        cmd.String("payload"),
			ContentType: cmd.String("content-type"),
		}
	}

	// aliases
	if cmd.IsSet("group") && input.Action == impl.ANone {
		// TODO We should support --group with --offline/--no-upload
		// TODO Ensure this works with --output-dir and --output-path
		input.Action = impl.ARunModule
		input.RunModuleArgs = impl.ARunModuleArgs{Name: []string{"advisor", "collect"}, Options: []string{"--group " + cmd.String("group")}}
	}
	if cmd.IsSet("compliance") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.RunModuleArgs = impl.ARunModuleArgs{Name: []string{"compliance", "collect"}}
	}
	if cmd.IsSet("check-results") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.RunModuleArgs = impl.ARunModuleArgs{Name: []string{"advisor", "check-results"}}
	}
	if cmd.IsSet("show-results") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.RunModuleArgs = impl.ARunModuleArgs{Name: []string{"advisor", "show-results"}}
	}

	if input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.RunModuleArgs = impl.ARunModuleArgs{Name: []string{"advisor", "collect"}, Options: parseModuleOptions(cmd.StringSlice("module-options"))}
	}

	// module flags
	if input.Action == impl.ARunModule {
		input.RunModuleArgs.OutputFile = filepath.Join(internal.ArchiveDirectoryParentPath, fmt.Sprintf("archive-%d.tar.xz", time.Now().Unix()))
		input.RunModuleArgs.OutputDir = cmd.String("output-dir")
		input.RunModuleArgs.OutputFile = cmd.String("output-file")
		input.RunModuleArgs.Options = append(input.RunModuleArgs.Options, parseModuleOptions(cmd.StringSlice("module-option"))...)
	}

	return input
}

func parseModuleOptions(options []string) []string {
	result := make([]string, 0)
	for _, option := range options {
		result = append(result, "--"+option)
	}
	return result
}

func runCLI(_ context.Context, cmd *cli.Command) error {
	if err := validateCLI(cmd); err != nil {
		return err
	}

	input := parseCLI(cmd)

	if input.Action == impl.AHelp {
		_ = cli.ShowAppHelp(cmd)
		return nil
	}

	// ask for elevated privileges
	if os.Geteuid() != 0 {
		return internal.NewError(internal.ErrPermissions, nil, "This command has to be run with superuser privileges.")
	}

	switch input.Action {
	case impl.ASetDisplayName:
		return impl.RunSetDisplayName(input)
	case impl.ASetAnsibleHostname:
		return impl.RunSetAnsibleHostname(input)
	case impl.ARegister:
		return impl.RunRegister(input)
	case impl.AUnregister:
		return impl.RunUnregister(input)
	case impl.AStatus:
		return impl.RunStatus(input)
	case impl.AListModules:
		return impl.RunListModules(input)
	case impl.ARunModule:
		return nil // Call module, maybe upload archive, and exit.
	case impl.AUploadLocalArchive:
		return nil // Upload archive and exit.
	default:
		return internal.NewError(internal.ErrInput, fmt.Errorf("bad input: %#v", input), "Not implemented.")
	}
}
