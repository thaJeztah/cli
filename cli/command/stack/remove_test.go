package stack

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/cli/internal/test"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func fakeClientForRemoveStackTest(version string) *fakeClient {
	allServices := []string{
		objectName("foo", "service1"),
		objectName("foo", "service2"),
		objectName("bar", "service1"),
		objectName("bar", "service2"),
	}
	allNetworks := []string{
		objectName("foo", "network1"),
		objectName("bar", "network1"),
	}
	allSecrets := []string{
		objectName("foo", "secret1"),
		objectName("foo", "secret2"),
		objectName("bar", "secret1"),
	}
	allConfigs := []string{
		objectName("foo", "config1"),
		objectName("foo", "config2"),
		objectName("bar", "config1"),
	}
	allVolumes := []string{
		objectName("foo", "volume1"),
		objectName("foo", "volume2"),
		objectName("bar", "volume1"),
	}
	return &fakeClient{
		version:  version,
		services: allServices,
		networks: allNetworks,
		secrets:  allSecrets,
		configs:  allConfigs,
		volumes:  allVolumes,
	}
}

func TestRemoveWithEmptyName(t *testing.T) {
	cmd := newRemoveCommand(test.NewFakeCli(&fakeClient{}), &orchestrator)
	cmd.SetArgs([]string{"good", "'   '", "alsogood"})
	cmd.SetOutput(ioutil.Discard)

	assert.ErrorContains(t, cmd.Execute(), `invalid stack name: "'   '"`)
}

func TestRemoveStackVersion124DoesNotRemoveConfigsOrSecrets(t *testing.T) {
	client := fakeClientForRemoveStackTest("1.24")
	cmd := newRemoveCommand(test.NewFakeCli(client), &orchestrator)
	cmd.SetArgs([]string{"foo", "bar"})

	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.services), client.removedServices))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.networks), client.removedNetworks))
	assert.Check(t, is.Len(client.removedSecrets, 0))
	assert.Check(t, is.Len(client.removedConfigs, 0))
}

func TestRemoveStackVersion125DoesNotRemoveConfigs(t *testing.T) {
	client := fakeClientForRemoveStackTest("1.25")
	cmd := newRemoveCommand(test.NewFakeCli(client), &orchestrator)
	cmd.SetArgs([]string{"foo", "bar"})

	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.services), client.removedServices))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.networks), client.removedNetworks))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.secrets), client.removedSecrets))
	assert.Check(t, is.Len(client.removedConfigs, 0))
}

func TestRemoveStackVersion130RemovesEverything(t *testing.T) {
	client := fakeClientForRemoveStackTest("1.30")
	cmd := newRemoveCommand(test.NewFakeCli(client), &orchestrator)
	cmd.SetArgs([]string{"foo", "bar"})

	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.services), client.removedServices))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.networks), client.removedNetworks))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.secrets), client.removedSecrets))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.configs), client.removedConfigs))
}

func TestRemoveStack130PreservesVolumes(t *testing.T) {
	client := fakeClientForRemoveStackTest("1.30")
	cmd := newRemoveCommand(test.NewFakeCli(client), &orchestrator)
	cmd.SetArgs([]string{"foo", "bar"})

	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.services), client.removedServices))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.networks), client.removedNetworks))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.secrets), client.removedSecrets))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.configs), client.removedConfigs))
	assert.Check(t, is.Len(client.removedVolumes, 0))
}

func TestRemoveStack130RemoveVolumes(t *testing.T) {
	client := fakeClientForRemoveStackTest("1.30")
	cmd := newRemoveCommand(test.NewFakeCli(client), &orchestrator)
	cmd.SetArgs([]string{"-v", "foo", "bar"})

	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.services), client.removedServices))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.networks), client.removedNetworks))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.secrets), client.removedSecrets))
	assert.Check(t, is.DeepEqual(buildObjectIDs(client.configs), client.removedConfigs))
	assert.Check(t, is.DeepEqual(client.volumes, client.removedVolumes))
}

func TestRemoveStackSkipEmpty(t *testing.T) {
	allServices := []string{objectName("bar", "service1"), objectName("bar", "service2")}
	allServiceIDs := buildObjectIDs(allServices)

	allNetworks := []string{objectName("bar", "network1")}
	allNetworkIDs := buildObjectIDs(allNetworks)

	allSecrets := []string{objectName("bar", "secret1")}
	allSecretIDs := buildObjectIDs(allSecrets)

	allConfigs := []string{objectName("bar", "config1")}
	allConfigIDs := buildObjectIDs(allConfigs)

	allVolumes := []string{objectName("bar", "volume1")}
	allVolumeIDs := allVolumes

	fakeClient := &fakeClient{
		version:  "1.30",
		services: allServices,
		networks: allNetworks,
		secrets:  allSecrets,
		configs:  allConfigs,
		volumes:  allVolumes,
	}
	fakeCli := test.NewFakeCli(fakeClient)
	cmd := newRemoveCommand(fakeCli, &orchestrator)
	cmd.SetArgs([]string{"-v", "foo", "bar"})

	assert.NilError(t, cmd.Execute())
	expectedList := []string{
		"Removing service bar_service1",
		"Removing service bar_service2",
		"Removing secret bar_secret1",
		"Removing config bar_config1",
		"Removing network bar_network1",
		"Removing volume bar_volume1",
	}
	assert.Check(t, is.Equal(strings.Join(expectedList, "\n")+"\n", fakeCli.OutBuffer().String()))
	assert.Check(t, is.Contains(fakeCli.ErrBuffer().String(), "Nothing found in stack: foo\n"))
	assert.Check(t, is.DeepEqual(allServiceIDs, fakeClient.removedServices))
	assert.Check(t, is.DeepEqual(allNetworkIDs, fakeClient.removedNetworks))
	assert.Check(t, is.DeepEqual(allSecretIDs, fakeClient.removedSecrets))
	assert.Check(t, is.DeepEqual(allConfigIDs, fakeClient.removedConfigs))
	assert.Check(t, is.DeepEqual(allVolumeIDs, fakeClient.removedVolumes))
}

func TestRemoveContinueAfterError(t *testing.T) {
	allServices := []string{objectName("foo", "service1"), objectName("bar", "service1")}
	allServiceIDs := buildObjectIDs(allServices)

	allNetworks := []string{objectName("foo", "network1"), objectName("bar", "network1")}
	allNetworkIDs := buildObjectIDs(allNetworks)

	allSecrets := []string{objectName("foo", "secret1"), objectName("bar", "secret1")}
	allSecretIDs := buildObjectIDs(allSecrets)

	allConfigs := []string{objectName("foo", "config1"), objectName("bar", "config1")}
	allConfigIDs := buildObjectIDs(allConfigs)

	allVolumes := []string{objectName("foo", "volume1"), objectName("bar", "volume1")}
	allVolumeIDs := allVolumes

	removedServices := []string{}
	cli := &fakeClient{
		version:  "1.30",
		services: allServices,
		networks: allNetworks,
		secrets:  allSecrets,
		configs:  allConfigs,
		volumes:  allVolumes,

		serviceRemoveFunc: func(serviceID string) error {
			removedServices = append(removedServices, serviceID)

			if strings.Contains(serviceID, "foo") {
				return errors.New("")
			}
			return nil
		},
	}
	cmd := newRemoveCommand(test.NewFakeCli(cli), &orchestrator)
	cmd.SetOutput(ioutil.Discard)
	cmd.SetArgs([]string{"-v", "foo", "bar"})

	assert.Error(t, cmd.Execute(), "Failed to remove some resources from stack: foo")
	assert.Check(t, is.DeepEqual(allServiceIDs, removedServices))
	assert.Check(t, is.DeepEqual(allNetworkIDs, cli.removedNetworks))
	assert.Check(t, is.DeepEqual(allSecretIDs, cli.removedSecrets))
	assert.Check(t, is.DeepEqual(allConfigIDs, cli.removedConfigs))
	assert.Check(t, is.DeepEqual(allVolumeIDs, cli.removedVolumes))
}
