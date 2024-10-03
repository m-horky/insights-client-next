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

// TODO Do not expose any new API.
// TODO Map everything into the pretty Module style objects on background.
// TODO Ensure the validation and parsing works for everything correctly.
// TODO Add positive and negative tests for the CLI.

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

// Flag is a proxy for cli.Flag object.
type Flag struct {
	Category string
	Type     rune
	Name     string
	Help     string
	Aliases  []string
}

// cliRootFlags defines all existing CLI flags.
var cliRootFlags = []Flag{
	{"HOST", 'b', "register", "register the host", []string{}},
	{"HOST", 'b', "unregister", "unregister the host", []string{}},
	{"HOST", 'b', "status", "display host status", []string{}},
	{"HOST", 'b', "checkin", "send lightweight check-in notification", []string{}},
	{"HOST", 'b', "test-connection", "test API connectivity", []string{}},
	{"HOST", 'b', "support", "generate data for customer support", []string{}},
	{"INVENTORY", 's', "display-name", "set display name of a host", []string{}},
	{"INVENTORY", 's', "ansible-host", "set Ansible display name of a host", []string{}},
	{"INVENTORY", 's', "group", "add system to Inventory group", []string{}},
	{"MODULES", 's', "output-dir", "do not upload, collect into directory", []string{}},
	{"MODULES", 's', "output-file", "do not upload, collect into file", []string{}},
	{"MODULES", 's', "payload", "upload archive from this path", []string{}},
	{"MODULES", 's', "content-type", "upload archive with this content type", []string{}},
	{"MODULES", 's', "collector", "run module collector", []string{"m"}},
	{"MODULES", 'b', "check-results", "download Advisor report", []string{}},
	{"MODULES", 'b', "show-results", "display Advisor report", []string{}},
	{"MODULES", 'b', "list-specs", "display Advisor collection specs", []string{}},
	{"MODULES", 'b', "diagnosis", "display Remediations report", []string{}},
	{"MODULES", 'b', "compliance", "run compliance", []string{}},
	{"MODULES", 'b', "no-upload", "alias for '--output-file [PATH]'", []string{}},
	{"MODULES", 'b', "keep-archive", "alias for '--output-file [PATH]'", []string{}},
	{"GLOBAL", 's', "format", "change output format", []string{}},
	{"GLOBAL", 'b', "debug", "print logs to stderr instead of a log file", []string{}},
	{"GLOBAL", 'b', "offline", "for some commands, only do local changes", []string{}},
	{"DEPRECATED", 'b', "retry", "ignored", []string{}},
	{"DEPRECATED", 'b', "validate", "ignored", []string{}},
	{"DEPRECATED", 'b', "quiet", "ignored", []string{}},
	{"DEPRECATED", 'b', "silent", "ignored", []string{}},
	{"DEPRECATED", 'b', "conf", "ignored", []string{"c"}},
	{"DEPRECATED", 'b', "compressor", "ignored", []string{}},
	{"DEPRECATED", 'b', "logging-file", "ignored", []string{}},
	{"DEPRECATED", 'b', "enable-schedule", "alias for '--register'", []string{}},
	{"DEPRECATED", 'b', "disable-schedule", "alias for '--unregister'", []string{}},
}

func buildHelpText() string {
	// FIXME Can we make this not break in narrow terminals?
	help := []string{`Usage: {{.Name}} [COMMAND] [FLAGS]`}

	maxFlagLength := 0
	for _, flag := range cliRootFlags {
		if len(buildHelpFlag(flag)) > maxFlagLength {
			maxFlagLength = len(buildHelpFlag(flag))
		}
	}
	lastCategory := ""
	for _, flag := range cliRootFlags {
		if flag.Category != lastCategory {
			help = append(help, ``)
			help = append(help, fmt.Sprintf("Category: %s", flag.Category))
			lastCategory = flag.Category
		}

		// left-justify flag names
		help = append(help, fmt.Sprintf(
			"  %s%s  %s",
			buildHelpFlag(flag),
			strings.Repeat(" ", maxFlagLength-len(buildHelpFlag(flag))),
			flag.Help,
		))
	}

	return strings.Join(help, "\n") + "\n"
}

// buildHelpFlag constructs a string out of the flag and its aliases
func buildHelpFlag(flag Flag) string {
	result := "--" + flag.Name
	for _, alias := range flag.Aliases {
		if len(alias) == 1 {
			result += ", -" + alias
		} else {
			result += ", --" + alias
		}
	}
	return result
}

func buildCLI() *cli.Command {
	var cliFlags []cli.Flag
	for _, flag := range cliRootFlags {
		switch flag.Type {
		case 'b':
			cliFlags = append(cliFlags, &cli.BoolFlag{Name: flag.Name, Aliases: flag.Aliases})
		case 's':
			cliFlags = append(cliFlags, &cli.StringFlag{Name: flag.Name, Aliases: flag.Aliases})
		case 'l':
			cliFlags = append(cliFlags, &cli.StringSliceFlag{Name: flag.Name, Aliases: flag.Aliases})
		default:
			panic(fmt.Sprintf("Unsupported flag type: %v", flag.Type))
		}
	}

	return &cli.Command{
		Name:            "insights-client",
		HideHelpCommand: true,
		Version:         internal.Version,
		Usage:           "Upload data to Red Hat Insights",
		UsageText:       fmt.Sprintf("%s COMMAND [FLAGS...]", "insights-client"),
		Flags:           cliFlags,
		Action:          runCLI,
	}
}

// validateCLI performs input validation.
//
// It shows notices for cliRootFlags that are deprecated.
//
// It ensures cliRootFlags that assume other cliRootFlags are properly joined.
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
		{"compliance", "output-dir"},
		{"compliance", "output-file"},
		{"compliance", "module-option", "output-dir"},
		{"compliance", "module-option", "output-file"},
		{"malware", "output-dir"},
		{"malware", "output-file"},
		{"malware", "module-option", "output-dir"},
		{"malware", "module-option", "output-file"},
		{"payload", "content-type"},
		{"test-connection"},
		{"support"},
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

			// we don't need to check global cliRootFlags, they can be applied to everything
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
				// resolved cliRootFlags conflict (e.g. `--compliance --diagnosis`)
				return internal.NewError(internal.ErrInput, errors.New("found conflict in module flags"), "This flag combination is not valid.")
			}
			setFlags[resolvedFlagName] = true
		}
	}

	// Exit immediately if no cliRootFlags were entered. Global cliRootFlags are not considered.
	if len(setFlags) == 0 {
		return nil
	}
	// Exit immediately if we find combination match: validation is complete. Global cliRootFlags are not considered.
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
	//  "Did you mean...?"

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
func parseCLI(cmd *cli.Command) (*impl.Input, error) {
	input := &impl.Input{}

	if cmd.IsSet("format") {
		format, err := internal.ParseFormat(cmd.String("format"))
		if err != nil {
			return nil, err
		}
		input.Format = format
	} else {
		input.Format = internal.Human
	}

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
	if cmd.IsSet("test-connection") && input.Action == impl.ANone {
		input.Action = impl.ATestConnection
	}
	if cmd.IsSet("support") && input.Action == impl.ANone {
		input.Action = impl.ASupport
	}

	// inventory
	if cmd.IsSet("display-name") && input.Action == impl.ANone {
		input.Action = impl.ASetDisplayName
		input.SetDisplayNameArgs = impl.ASetDisplayNameArgs{Name: cmd.String("display-name")}
	}
	if cmd.IsSet("ansible-host") && input.Action == impl.ANone {
		input.Action = impl.ASetAnsibleHostname
		input.SetAnsibleHostnameArgs = impl.ASetAnsibleHostnameArgs{Name: cmd.String("ansible-host")}
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
	if cmd.IsSet("group") && cmd.IsSet("offline") && input.Action == impl.ANone {
		input.Action = impl.ASetGroupLocally
	}
	if cmd.IsSet("group") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.RunModuleArgs = impl.ARunModuleArgs{
			Name:    []string{"advisor", "collect"},
			Options: []string{"--group " + cmd.String("group")},
		}
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

	// module cliRootFlags
	if input.Action == impl.ARunModule {
		if !modules.CommandExists(input.RunModuleArgs.Name) {
			return nil, internal.NewError(nil, nil, fmt.Sprintf("No module implements command '%s'.", strings.Join(input.RunModuleArgs.Name, " ")))
		}

		input.RunModuleArgs.OutputFile = filepath.Join(internal.ArchiveDirectoryParentPath, fmt.Sprintf("archive-%d.tar.xz", time.Now().Unix()))
		input.RunModuleArgs.OutputDir = cmd.String("output-dir")
		input.RunModuleArgs.OutputFile = cmd.String("output-file")
		input.RunModuleArgs.Options = append(input.RunModuleArgs.Options, parseModuleOptions(cmd.StringSlice("module-option"))...)
	}

	return input, nil
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

	input, err := parseCLI(cmd)
	if err != nil {
		return err
	}

	if input.Action == impl.AHelp {
		_ = cli.ShowAppHelp(cmd)
		return nil
	}

	// ask for elevated privileges
	if os.Geteuid() != 0 {
		return internal.NewError(internal.ErrPermissions, nil, "This command has to be run with superuser privileges.")
	}

	switch input.Action {
	case impl.ARegister:
		return impl.RunRegister(input)
	case impl.AUnregister:
		return impl.RunUnregister(input)
	case impl.AStatus:
		return impl.RunStatus(input)
	case impl.ACheckIn:
		return impl.RunCheckIn(input)
	case impl.ASetDisplayName:
		return impl.RunSetDisplayName(input)
	case impl.ASetAnsibleHostname:
		return impl.RunSetAnsibleHostname(input)
	case impl.ARunModule:
		// TODO validate module first
		return impl.RunModule(input)
	case impl.AListModules:
		return impl.RunListModules(input)
	case impl.AUploadLocalArchive:
		return impl.RunUploadLocalArchive(input)
	case impl.ATestConnection:
		return impl.RunTestConnection(input)
	case impl.ASupport:
		return impl.RunSupport(input)
	case impl.ASetGroupLocally:
		return impl.RunSetGroupLocally(input)
	default:
		return internal.NewError(internal.ErrInput, fmt.Errorf("bad input: %#v", input), "Not implemented.")
	}
}
