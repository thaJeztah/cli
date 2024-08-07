package cli

import (
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
)

func TestRequiresNoArgs(t *testing.T) {
	testCases := []testCase{
		{
			validateFunc:  NoArgs,
			expectedError: "no error",
		},
		{
			args:          []string{"foo"},
			validateFunc:  NoArgs,
			expectedError: "accepts no arguments",
		},
	}
	for _, tc := range testCases {
		cmd := newDummyCommand(tc.validateFunc)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestRequiresMinArgs(t *testing.T) {
	testCases := []testCase{
		{
			validateFunc:  RequiresMinArgs(0),
			expectedError: "no error",
		},
		{
			validateFunc:  RequiresMinArgs(1),
			expectedError: "at least 1 argument",
		},
		{
			args:          []string{"foo"},
			validateFunc:  RequiresMinArgs(2),
			expectedError: "at least 2 arguments",
		},
	}
	for _, tc := range testCases {
		cmd := newDummyCommand(tc.validateFunc)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestRequiresMaxArgs(t *testing.T) {
	testCases := []testCase{
		{
			validateFunc:  RequiresMaxArgs(0),
			expectedError: "no error",
		},
		{
			args:          []string{"foo", "bar"},
			validateFunc:  RequiresMaxArgs(1),
			expectedError: "at most 1 argument",
		},
		{
			args:          []string{"foo", "bar", "baz"},
			validateFunc:  RequiresMaxArgs(2),
			expectedError: "at most 2 arguments",
		},
	}
	for _, tc := range testCases {
		cmd := newDummyCommand(tc.validateFunc)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestRequiresRangeArgs(t *testing.T) {
	testCases := []testCase{
		{
			validateFunc:  RequiresRangeArgs(0, 0),
			expectedError: "no error",
		},
		{
			validateFunc:  RequiresRangeArgs(0, 1),
			expectedError: "no error",
		},
		{
			args:          []string{"foo", "bar"},
			validateFunc:  RequiresRangeArgs(0, 1),
			expectedError: "at most 1 argument",
		},
		{
			args:          []string{"foo", "bar", "baz"},
			validateFunc:  RequiresRangeArgs(0, 2),
			expectedError: "at most 2 arguments",
		},
		{
			validateFunc:  RequiresRangeArgs(1, 2),
			expectedError: "at least 1 ",
		},
	}
	for _, tc := range testCases {
		cmd := newDummyCommand(tc.validateFunc)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestExactArgs(t *testing.T) {
	testCases := []testCase{
		{
			validateFunc:  ExactArgs(0),
			expectedError: "no error",
		},
		{
			validateFunc:  ExactArgs(1),
			expectedError: "1 argument",
		},
		{
			validateFunc:  ExactArgs(2),
			expectedError: "2 arguments",
		},
	}
	for _, tc := range testCases {
		cmd := newDummyCommand(tc.validateFunc)
		cmd.SetArgs(tc.args)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

type testCase struct {
	args          []string
	validateFunc  cobra.PositionalArgs
	expectedError string
}

func newDummyCommand(validationFunc cobra.PositionalArgs) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "dummy",
		Args: validationFunc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("no error")
		},
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd
}
