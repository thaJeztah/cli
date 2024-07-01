package manager

import (
	"fmt"
	"strings"

	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/version"
	"github.com/shirou/gopsutil/v3/host"
)

const (
	audience = "https://hub.docker.com"
	tenant   = "login.docker.com"
	clientID = "DHWuMefQ1v4lxENpz8oUYH50yYSwyPvi"
)

func NewManager(store credentials.Store) (*OAuthManager, error) {
	hostinfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	cliVersion := strings.ReplaceAll(version.Version, ".", "_")
	options := OAuthManagerOptions{
		Audience:   audience,
		ClientID:   clientID,
		Tenant:     tenant,
		DeviceName: "docker-cli:" + cliVersion,
		Store:      store,
	}

	if hostinfo != nil {
		hostVersion := strings.ReplaceAll(hostinfo.PlatformVersion, ".", "_")
		options.DeviceName = fmt.Sprintf("docker-cli:%s:%s-%s-%s", cliVersion, hostinfo.OS, hostVersion, hostinfo.KernelArch)
	}

	authManager, err := New(options)
	if err != nil {
		return nil, err
	}
	return authManager, nil
}
