package cmd

import (
	"io/ioutil"

	"code.cloudfoundry.org/cfdev/config"
	"github.com/spf13/cobra"
)

func NewBosh(Exit chan struct{}, UI UI, Config config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "bosh",
		Run: func(cmd *cobra.Command, args []string) {
			UI.Say(`Usage: eval $(cf dev bosh env)`)
		},
	}
	envCmd := &cobra.Command{
		Use: "env",
		RunE: func(cmd *cobra.Command, args []string) error {
			shellScript, err := ioutil.ReadFile("/var/vcap/director/env")
			if err != nil {
				return err
			}
			UI.Say(string(shellScript))
			return nil
		},
	}
	cmd.AddCommand(envCmd)
	return cmd
}
