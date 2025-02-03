package ssh

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestParseURL(t *testing.T) {
	testCases := []struct {
		url           string
		expectedArgs  []string
		expectedError string
		expectedSpec  Spec
	}{
		{
			url: "ssh://example.com",
			expectedArgs: []string{
				"--", "example.com",
			},
			expectedSpec: Spec{
				Host: "example.com",
			},
		},
		{
			url: "ssh://me@example.com:10022",
			expectedArgs: []string{
				"-l", "me",
				"-p", "10022",
				"--", "example.com",
			},
			expectedSpec: Spec{
				User: "me",
				Host: "example.com",
				Port: "10022",
			},
		},
		{
			url: "ssh://me@example.com:10022/var/run/docker.sock",
			expectedArgs: []string{
				"-l", "me",
				"-p", "10022",
				"--", "example.com",
			},
			expectedSpec: Spec{
				User: "me",
				Host: "example.com",
				Port: "10022",
				Path: "/var/run/docker.sock",
			},
		},
		{
			url:           "ssh://me:passw0rd@example.com",
			expectedError: "plain-text password is not supported",
		},
		{
			url:           "ssh://example.com?bar",
			expectedError: `extra query after the host: "bar"`,
		},
		{
			url:           "ssh://example.com#bar",
			expectedError: `extra fragment after the host: "bar"`,
		},
		{
			url:           "ssh://",
			expectedError: "no host specified",
		},
		{
			url:           "foo://example.com",
			expectedError: `expected scheme ssh, got "foo"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			sp, err := ParseURL(tc.url)
			if tc.expectedError == "" {
				assert.NilError(t, err)
				assert.Check(t, is.DeepEqual(sp.Args(), tc.expectedArgs))
				assert.Check(t, is.Equal(*sp, tc.expectedSpec))
			} else {
				assert.Check(t, is.Error(err, tc.expectedError))
				assert.Check(t, is.Nil(sp))
			}
		})
	}
}
