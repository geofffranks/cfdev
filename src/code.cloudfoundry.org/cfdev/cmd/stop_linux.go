package cmd

import (
	"fmt"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/spf13/cobra"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

func NewStop(Config *config.Config, AnalyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := process.SignalAndCleanup(filepath.Join(Config.CFDevHome, "garden.pid"), "/var/vcap/gdn.socket", syscall.SIGTERM); err != nil {
				return fmt.Errorf("try using sudo ; failed to terminate garden: %s", err)
			}
			return nil
		},
	}
	return cmd
}
