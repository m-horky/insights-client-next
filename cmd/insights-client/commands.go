package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
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

		// write empty line to the log, to make new invocations more obvious
		_, _ = fp.WriteString("\n")
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
	{"COLLECTION", 's', "output-dir", "do not upload, collect into directory", []string{}},
	{"COLLECTION", 's', "output-file", "do not upload, collect into file", []string{}},
	{"COLLECTION", 's', "payload", "upload archive from this path", []string{}},
	{"COLLECTION", 's', "content-type", "upload archive with this content type", []string{}},
	{"COLLECTION", 's', "collector", "run module collector", []string{"m"}},
	{"COLLECTION", 'b', "check-results", "download Advisor report", []string{}},
	{"COLLECTION", 'b', "show-results", "display Advisor report", []string{}},
	{"COLLECTION", 'b', "list-specs", "display Advisor collection specs", []string{}},
	{"COLLECTION", 'b', "diagnosis", "display Remediations report", []string{}},
	{"COLLECTION", 'b', "compliance", "run compliance", []string{}},
	{"COLLECTION", 'b', "no-upload", "alias for '--output-file [PATH]'", []string{}},
	{"COLLECTION", 'b', "keep-archive", "alias for '--output-file [PATH]'", []string{}},
	{"COLLECTION", 's', "manifest", "run Advisor with manifest", []string{}},
	{"COLLECTION", 'b', "validate", "validate Advisor denylist", []string{}},
	{"COLLECTION", 's', "build-packagecache", "refresh system package manager cache", []string{}},
	{"GLOBAL", 's', "format", "change output format", []string{}},
	{"GLOBAL", 'b', "debug", "print logs to stderr instead of a log file", []string{}},
	{"GLOBAL", 'b', "offline", "for some commands, only do local changes", []string{}},
	{"DEPRECATED", 's', "retry", "ignored", []string{}},
	{"DEPRECATED", 'b', "quiet", "ignored", []string{}},
	{"DEPRECATED", 'b', "silent", "ignored", []string{}},
	{"DEPRECATED", 'b', "conf", "ignored", []string{"c"}},
	{"DEPRECATED", 'b', "compressor", "ignored", []string{}},
	{"DEPRECATED", 'b', "logging-file", "ignored", []string{}},
	{"DEPRECATED", 'b', "net-debug", "ignored", []string{}},
	{"DEPRECATED", 'b', "enable-schedule", "alias for '--register'", []string{}},
	{"DEPRECATED", 'b', "disable-schedule", "alias for '--unregister'", []string{}},
}

func buildHelpText() string {
	// FIXME Can we make this not break in narrow terminals?
	help := []string{`Usage: insights-client [COMMAND] [FLAGS]`}

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
	noopFlags := []string{"quiet", "retry", "silent", "conf", "compressor", "logging-file", "net-debug"}

	// this includes the list of all valid combinations
	flagCombinations := [][]string{
		// HOST
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
		{"test-connection"},
		{"support"},
		// INVENTORY
		{"display-name"},
		{"ansible-host"},
		{"group"},
		{"group", "offline"},
		// COLLECTION
		{"payload", "content-type"},
		{"output-dir"},
		{"output-file"},
		{"collector"},
		{"collector", "output-dir"},
		{"collector", "output-file"},
		{"collector", "no-upload"},
		{"collector", "keep-archive"},
		{"collector", "offline"},
		{"compliance"},
		{"compliance", "output-dir"},
		{"compliance", "output-file"},
		{"compliance", "no-upload"},
		{"compliance", "keep-archive"},
		{"compliance", "offline"},
		{"offline"},
		{"check-results"},
		{"show-results"},
		{"list-specs"},
		{"diagnosis"},
		{"no-upload"},
		{"keep-archive"},
		{"manifest"},
		{"build-packagecache"},
		// DEPRECATED
		{"validate"},
		{"enable-schedule"},
		{"disable-schedule"},
	}

	setFlags := make(map[string]bool)

	for _, flag := range cmd.Flags {
		flagName := flag.Names()[0]

		if cmd.IsSet(flagName) {
			// we don't need to check global flags, they can be applied to everything
			flagIsGlobal := false
			for _, globalFlag := range globalFlags {
				if flagName == globalFlag {
					flagIsGlobal = true
					break
				}
			}
			for _, globalFlag := range noopFlags {
				if flagName == globalFlag {
					flagIsGlobal = true
					break
				}
			}
			if flagIsGlobal {
				continue
			}

			setFlags[flagName] = true
		}
	}

	// Exit immediately if no flags were entered. Global flags are not considered.
	if len(setFlags) == 0 {
		return nil
	}
	// Exit immediately if we find combination match: validation is complete.
	var finalFlags []string
	for flag := range setFlags {
		finalFlags = append(finalFlags, flag)
	}
	sort.Strings(finalFlags)
	for _, combination := range flagCombinations {
		sort.Strings(combination)
		if reflect.DeepEqual(combination, finalFlags) {
			return nil
		}
	}

	// TODO Display all flag combinations that use the entered flags
	//  "Did you mean...?"

	setFlagList := make([]string, 0)
	for flag, _ := range setFlags {
		setFlagList = append(setFlagList, flag)
	}
	return internal.NewError(
		internal.ErrInput,
		fmt.Errorf("bad flag combination: %s", strings.Join(setFlagList, ",")),
		"This flag combination is not valid.",
	)
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
		input.Args = impl.ARegisterArgs{
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
		input.Args = impl.ASetDisplayNameArgs{Name: cmd.String("display-name")}
	}
	if cmd.IsSet("ansible-host") && input.Action == impl.ANone {
		input.Action = impl.ASetAnsibleHostname
		input.Args = impl.ASetAnsibleHostnameArgs{Name: cmd.String("ansible-host")}
	}
	if cmd.IsSet("group") && cmd.IsSet("offline") && input.Action == impl.ANone {
		// TODO Should we check that it is not not empty string?
		input.Action = impl.ASetGroupLocally
		input.Args = impl.ASetGroupLocallyArgs{Name: cmd.String("group")}
	}
	if cmd.IsSet("group") && input.Action == impl.ANone {
		// TODO Should we check that it is not not empty string?
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{
			Command: modules.GetAdvisorModule().ArchiveCommandName,
			Options: []string{"--group", cmd.String("group")},
		}
	}

	// collection
	if cmd.IsSet("payload") && cmd.IsSet("content-type") && input.Action == impl.ANone {
		input.Action = impl.AUploadLocalArchive
		input.Args = impl.AUploadLocalArchiveArgs{
			Path:        cmd.String("payload"),
			ContentType: cmd.String("content-type"),
		}
	}
	if cmd.IsSet("collector") && input.Action == impl.ANone {
		switch cmd.String("collector") {
		case "malware-detection":
			input.Action = impl.ARunModule
			input.Args = impl.ARunModuleArgs{Command: modules.GetMalwareModule().ArchiveCommandName}
		default:
			return nil, internal.NewError(nil, nil, fmt.Sprintf("Collector not known: '%s'.", cmd.String("collector")))
		}
	}
	if cmd.IsSet("compliance") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: modules.GetComplianceModule().ArchiveCommandName}
	}
	if cmd.IsSet("check-results") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: []string{"advisor", "check-results"}}
	}
	if cmd.IsSet("show-results") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: []string{"advisor", "show-results"}}
	}
	if cmd.IsSet("list-specs") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: []string{"advisor", "list-specs"}}
	}
	if cmd.IsSet("diagnosis") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: []string{"advisor", "diagnosis"}}
	}
	if cmd.IsSet("manifest") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: []string{"advisor", "manifest"}}
	}
	if cmd.IsSet("build-packagecache") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: []string{"advisor", "build-packagecache"}}
	}
	if cmd.IsSet("validate") && input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: []string{"advisor", "validate"}}
	}

	// default action
	if input.Action == impl.ANone {
		input.Action = impl.ARunModule
		input.Args = impl.ARunModuleArgs{Command: modules.GetAdvisorModule().ArchiveCommandName}
	}

	// module flags
	if input.Action == impl.ARunModule {
		args := input.Args.(impl.ARunModuleArgs)
		if !modules.CommandExists(args.Command) {
			return nil, internal.NewError(nil, nil, fmt.Sprintf("No module implements command '%s'.", strings.Join(args.Command, " ")))
		}

		// We have to figure out what to do based on the following options:
		// --no-upload
		// --keep-archive
		// --offline
		// --output-dir
		// --output-file
		if cmd.IsSet("output-dir") {
			args.ArchiveParent = cmd.String("output-dir")
			args.ArchiveName = fmt.Sprintf("archive-%d", time.Now().Unix())
			args.StopAtDir = true
		} else if cmd.IsSet("output-file") {
			args.ArchiveParent = filepath.Dir(cmd.String("output-file"))
			args.ArchiveName = strings.Split(filepath.Base(cmd.String("output-file")), ".")[0]
			args.StopAtFile = true
		} else if cmd.IsSet("no-upload") {
			args.ArchiveParent = internal.ArchiveDirectoryParentPath
			args.ArchiveName = fmt.Sprintf("archive-%d", time.Now().Unix())
			args.StopAtFile = true
		} else if cmd.IsSet("offline") {
			args.ArchiveParent = internal.ArchiveDirectoryParentPath
			args.ArchiveName = fmt.Sprintf("archive-%d", time.Now().Unix())
			args.StopAtFile = true
		} else if cmd.IsSet("keep-archive") {
			args.ArchiveParent = internal.ArchiveDirectoryParentPath
			args.ArchiveName = fmt.Sprintf("archive-%d", time.Now().Unix())
			args.StopAtCleanup = true
		}
		input.Args = args
	}

	return input, nil
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
		return impl.RunModule(input)
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
