package image

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/cli/internal/test"
	"github.com/moby/moby/api/types/image"
	"gotest.tools/v3/assert"
)

func TestNewPushCommandErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedError string
		imagePushFunc func(ref string, options image.PushOptions) (io.ReadCloser, error)
	}{
		{
			name:          "wrong-args",
			args:          []string{},
			expectedError: "requires 1 argument",
		},
		{
			name:          "invalid-name",
			args:          []string{"UPPERCASE_REPO"},
			expectedError: "invalid reference format: repository name (library/UPPERCASE_REPO) must be lowercase",
		},
		{
			name:          "push-failed",
			args:          []string{"image:repo"},
			expectedError: "Failed to push",
			imagePushFunc: func(ref string, options image.PushOptions) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("")), errors.New("Failed to push")
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := test.NewFakeCli(&fakeClient{imagePushFunc: tc.imagePushFunc})
			cmd := NewPushCommand(cli)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tc.args)
			assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
		})
	}
}

func TestNewPushCommandSuccess(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		output string
	}{
		{
			name: "push",
			args: []string{"image:tag"},
		},
		{
			name: "push quiet",
			args: []string{"--quiet", "image:tag"},
			output: `docker.io/library/image:tag
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := test.NewFakeCli(&fakeClient{
				imagePushFunc: func(ref string, options image.PushOptions) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
			})
			cmd := NewPushCommand(cli)
			cmd.SetOut(cli.OutBuffer())
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tc.args)
			assert.NilError(t, cmd.Execute())
			if tc.output != "" {
				assert.Equal(t, tc.output, cli.OutBuffer().String())
			}
		})
	}
}
