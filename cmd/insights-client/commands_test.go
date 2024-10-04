package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/m-horky/insights-client-next/internal/impl"
)

// TestValidateCLI_valid ensures that some flags can be used together.
//
// This is NOT an exhaustive list.
func TestValidateCLI_valid(t *testing.T) {
	tests := []struct {
		Input []string
	}{
		{[]string{""}},
		{[]string{"--register"}},
		{[]string{"--unregister"}},
		{[]string{"--register", "--group", "x"}},
		{[]string{"--register", "--display-name", "x"}},
		{[]string{"--register", "--display-name", "x", "--ansible-host", "x"}},
		{[]string{"--status"}},
		{[]string{"--output-file", "x"}},
		{[]string{"--output-dir", "x"}},
		{[]string{"--checkin"}},
		{[]string{"--payload", "x", "--content-type", "x"}},
		{[]string{"--compliance"}},
		{[]string{"--compliance", "--no-upload"}},
		{[]string{"--collector", "x"}},
		{[]string{"-m", "x"}},
		{[]string{"-m", "x", "--no-upload"}},
		{[]string{"-m", "x", "--keep-archive"}},
		{[]string{"-m", "x", "--output-file", "x"}},
		{[]string{"-m", "x", "--output-dir", "x"}},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.Input, " "), func(t *testing.T) {
			input := buildCLI()

			input.Action = func(_ context.Context, command *cli.Command) error {
				err := validateCLI(command)
				if err != nil {
					t.Errorf("expected 'nil', got '%v'", err)
				}
				return nil
			}

			args := []string{input.Name}
			args = append(args, test.Input...)
			_ = input.Run(context.Background(), args)
		})
	}
}

// TestValidateCLI_invalid ensures that some flags cannot be used together.
//
// This is NOT an exhaustive list.
func TestValidateCLI_invalid(t *testing.T) {
	tests := []struct {
		Input []string
	}{
		{[]string{"--register", "--unregister"}},
		{[]string{"--register", "--status"}},
		{[]string{"--test-connection", "--support"}},
		{[]string{"--compliance", "--display-name", "x"}},
		{[]string{"--status", "--content-type", "x"}},
		{[]string{"--payload", "x", "--offline"}},
		{[]string{"--payload", "x", "--output-dir", "x"}},
		{[]string{"-m", "x", "--display-name", "x"}},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.Input, " "), func(t *testing.T) {
			input := buildCLI()

			input.Action = func(_ context.Context, command *cli.Command) error {
				err := validateCLI(command)
				if err == nil {
					t.Errorf("expected error, got 'nil'")
				}
				return nil
			}

			args := []string{input.Name}
			args = append(args, test.Input...)
			_ = input.Run(context.Background(), args)
		})

	}
}

func TestParseCLI(t *testing.T) {
	tests := []struct {
		Input  []string
		Action impl.InputAction
		Args   any
	}{
		// host
		{[]string{"--register"}, impl.ARegister, impl.ARegisterArgs{}},
		{[]string{"--register", "--display-name", "x"}, impl.ARegister, impl.ARegisterArgs{DisplayName: "x"}},
		{[]string{"--register", "--ansible-host", "x"}, impl.ARegister, impl.ARegisterArgs{AnsibleHostname: "x"}},
		{[]string{"--register", "--group", "x"}, impl.ARegister, impl.ARegisterArgs{Group: "x"}},
		{[]string{"--unregister"}, impl.AUnregister, nil},
		{[]string{"--status"}, impl.AStatus, nil},
		{[]string{"--checkin"}, impl.ACheckIn, nil},
		{[]string{"--test-connection"}, impl.ATestConnection, nil},
		// inventory
		{[]string{"--display-name", "x"}, impl.ASetDisplayName, impl.ASetDisplayNameArgs{Name: "x"}},
		{[]string{"--ansible-host", "x"}, impl.ASetAnsibleHostname, impl.ASetAnsibleHostnameArgs{Name: "x"}},
		{[]string{"--group", "x", "--offline"}, impl.ASetGroupLocally, impl.ASetGroupLocallyArgs{Name: "x"}},
		{[]string{"--group", "x"}, impl.ARunModule, impl.ARunModuleArgs{
			Command:       []string{"advisor", "collect"},
			Options:       []string{"--group", "x"},
			ArchiveParent: "",
			ArchiveName:   "",
			StopAtDir:     false,
			StopAtFile:    false,
			StopAtCleanup: false,
		}},
		// collection
		{[]string{"--output-dir", "x"}, impl.ARunModule, impl.ARunModuleArgs{
			Command:       []string{"advisor", "collect"},
			Options:       nil,
			ArchiveParent: "x",
			ArchiveName:   fmt.Sprintf("archive-%d", time.Now().Unix()),
			StopAtDir:     true,
			StopAtFile:    false,
			StopAtCleanup: false,
		}},
		{[]string{"--compliance", "--output-file", "x"}, impl.ARunModule, impl.ARunModuleArgs{
			Command:       []string{"compliance", "collect"},
			Options:       nil,
			ArchiveParent: ".",
			ArchiveName:   "x",
			StopAtDir:     false,
			StopAtFile:    true,
			StopAtCleanup: false,
		}},
		{[]string{"--collector", "malware-detection", "--keep-archive"}, impl.ARunModule, impl.ARunModuleArgs{
			Command:       []string{"malware", "collect"},
			Options:       nil,
			ArchiveParent: "/var/cache/insights-client/",
			ArchiveName:   fmt.Sprintf("archive-%d", time.Now().Unix()),
			StopAtDir:     false,
			StopAtFile:    false,
			StopAtCleanup: true,
		}},
		// TODO Add more tests
		// {[]string{"--manifest", "x"}, impl.ARunModule, impl.ARunModuleArgs{}},
		// {[]string{"--build-packagecache"}, impl.ARunModule, impl.ARunModuleArgs{}},
		// {[]string{"--list-specs"}, impl.ARunModule, impl.ARunModuleArgs{}},
		// {[]string{"--validate"}, impl.ARunModule, impl.ARunModuleArgs{}},
		// {[]string{"--diagnosis"}, impl.ARunModule, impl.ARunModuleArgs{}},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.Input, " "), func(t *testing.T) {
			input := buildCLI()

			input.Action = func(_ context.Context, command *cli.Command) error {
				parsed, err := parseCLI(command)
				if err != nil {
					t.Fatalf("expected 'nil', got '%v'", err)
				}

				if parsed.Action != test.Action {
					t.Fatalf("expected '%v', got '%v'", test.Action, parsed.Action)
				}

				if !reflect.DeepEqual(test.Args, parsed.Args) {
					t.Fatalf("expected '%+v', got '%+v'", test.Args, parsed.Args)
				}
				return nil
			}

			args := []string{input.Name}
			args = append(args, test.Input...)
			_ = input.Run(context.Background(), args)
		})
	}
}
