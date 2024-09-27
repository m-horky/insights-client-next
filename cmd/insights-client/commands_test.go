package main

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestValidateCLI_valid ensures that some flags can be used together.
//
// This is NOT an exhaustive list.
func TestValidateCLI_valid(t *testing.T) {
	tests := []struct {
		Args []string
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
	}

	for _, test := range tests {
		input := buildCLI()

		input.Action = func(_ context.Context, command *cli.Command) error {
			err := validateCLI(command)
			if err != nil {
				t.Errorf("%v: expected 'nil', got '%v'", test.Args, err)
			}
			return nil
		}

		args := []string{input.Name}
		args = append(args, test.Args...)
		_ = input.Run(context.Background(), args)
	}
}

// TestValidateCLI_invalid ensures that some flags cannot be used together.
//
// This is NOT an exhaustive list.
func TestValidateCLI_invalid(t *testing.T) {
	tests := []struct {
		Args []string
	}{
		{[]string{"--register", "--unregister"}},
		{[]string{"--register", "--status"}},
		{[]string{"--test-connection", "--support"}},
		{[]string{"--module", "x", "--display-name"}},
		{[]string{"--status", "--content-type", "x"}},
		{[]string{"--payload", "x", "--offline"}},
		{[]string{"--payload", "x", "--output-dir", "x"}},
	}

	for _, test := range tests {
		input := buildCLI()

		input.Action = func(_ context.Context, command *cli.Command) error {
			err := validateCLI(command)
			if err == nil {
				t.Errorf("%v: expected error, got 'nil'", test.Args)
			}
			return nil
		}

		args := []string{input.Name}
		args = append(args, test.Args...)
		_ = input.Run(context.Background(), args)
	}
}
