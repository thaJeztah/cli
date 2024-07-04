package manager

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/version"
)

const (
	audience = "https://hub.docker.com"
	tenant   = "login.docker.com"
	clientID = "DHWuMefQ1v4lxENpz8oUYH50yYSwyPvi"
)

func NewManager(store credentials.Store) (*OAuthManager, error) {
	cliVersion := strings.ReplaceAll(version.Version, ".", "_")
	options := OAuthManagerOptions{
		Audience:   audience,
		ClientID:   clientID,
		Tenant:     tenant,
		DeviceName: "docker-cli:" + cliVersion,
		Store:      store,
	}

	// FIXME(thaJeztah): what information do we need here? Would https://github.com/moby/moby/blob/1205a9073320fba9e67fa5de3857f0330e56ce50/pkg/parsers/kernel/kernel_darwin.go#L13-L24 work?
	// hostVersion := strings.ReplaceAll(hostinfo.PlatformVersion, ".", "_")
	hostVersion := "unknown"
	// options.DeviceName = fmt.Sprintf("docker-cli:%s:%s-%s-%s", cliVersion, hostinfo.OS, hostVersion, hostinfo.KernelArch)
	options.DeviceName = fmt.Sprintf("docker-cli:%s:%s-%s-%s", cliVersion, runtime.GOOS, hostVersion, runtime.GOARCH)

	authManager, err := New(options)
	if err != nil {
		return nil, err
	}
	return authManager, nil
}
