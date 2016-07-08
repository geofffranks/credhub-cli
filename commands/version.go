package commands

import (
	"fmt"

	"os"

	"github.com/pivotal-cf/cm-cli/actions"
	"github.com/pivotal-cf/cm-cli/client"
	"github.com/pivotal-cf/cm-cli/config"
	"github.com/pivotal-cf/cm-cli/version"
)

func PrintVersion() error {
	cfg := config.ReadConfig()

	cmVersion := actions.NewVersion(client.NewHttpClient(cfg), cfg).GetServerVersion()

	fmt.Println("CLI Version:", version.Version+" build "+version.BuildNumber)
	fmt.Println("CM Version:", cmVersion)

	return nil
}

func init() {
	CM.Version = func() {
		PrintVersion()
		os.Exit(0)
	}
}
