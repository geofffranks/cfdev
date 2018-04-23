package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/garden/client"
	"github.com/spf13/cobra"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

type UI interface {
	Say(message string, args ...interface{})
}

type start struct {
	Exit            chan struct{}
	UI              UI
	Config          config.Config
	AnalyticsClient analytics.Client
	Registries      string
	gdnServer       *exec.Cmd
}

func NewStart(Exit chan struct{}, UI UI, Config config.Config, AnalyticsClient analytics.Client) *cobra.Command {
	s := start{Exit: Exit, UI: UI, Config: Config, AnalyticsClient: AnalyticsClient}
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.RunE()
		},
	}
	pf := cmd.PersistentFlags()
	pf.StringVar(&s.Registries, "r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")

	return cmd
}

func (s *start) RunE() error {
	go func() {
		<-s.Exit
		s.gdnServer.Process.Kill()
		os.Exit(128)
	}()

	cfanalytics.TrackEvent(cfanalytics.START_BEGIN, map[string]interface{}{"type": "cf"}, s.AnalyticsClient)

	if err := env.Setup(s.Config); err != nil {
		return err
	}

	garden := gdn.NewClient()
	// if garden.Ping() == nil {
	// 	s.UI.Say("CF Dev is already running...")
	// 	cfanalytics.TrackEvent(cfanalytics.START_END, map[string]interface{}{"type": "cf", "alreadyrunning": true}, s.AnalyticsClient)
	// 	return nil
	// }

	// TODO should this be the same on linux as on darwin????
	registries, err := s.parseDockerRegistriesFlag(s.Registries)
	if err != nil {
		return fmt.Errorf("Unable to parse docker registries %v\n", err)
	}

	s.UI.Say("Downloading Resources...")
	if err := download(s.Config.Dependencies, s.Config.CacheDir); err != nil {
		return err
	}

	// USER RUN
	// sudo iptables -A FORWARD -j ACCEPT

	s.UI.Say("Deploying the BOSH Director...")
	if err := gdn.DeployBosh(garden); err != nil {
		fmt.Printf("Failed to deploy the BOSH Director: %v\n", err)
		return fmt.Errorf("Failed to deploy the BOSH Director: %v\n", err)
	}

	if 2 == 2 {
		return nil
	}

	s.UI.Say("Deploying CF...")
	if err := gdn.DeployCloudFoundry(garden, registries); err != nil {
		return fmt.Errorf("Failed to deploy the Cloud Foundry: %v\n", err)
	}
	s.UI.Say(cfdevStartedMessage)

	cfanalytics.TrackEvent(cfanalytics.START_END, map[string]interface{}{"type": "cf"}, s.AnalyticsClient)

	return nil
}

// func (s *start) mountOssDepIso() error {
// 	u, err := user.Current()
// 	if err != nil {
// 		return fmt.Errorf("finding current user: %s", err)
// 	}
// 	isoPath := filepath.Join(s.Config.CFDevHome, "cache", "cf-oss-deps.iso")
// 	exec.Command("sudo", "umount", "/var/vcap/cache").Run() // TODO maybe not????
// 	os.MkdirAll("/var/vcap/cache", 0755)
// 	cmd := exec.Command("sudo", "mount", "-o", "loop,uid="+u.Uid, isoPath, "/var/vcap/cache")
// 	if out, err := cmd.CombinedOutput(); err != nil {
// 		os.Stdout.Write(out)
// 		return fmt.Errorf("mounting %s as %s: %s", isoPath, u.Uid, err)
// 	}
// 	return nil
// }
//
// func (s *start) unMountOssDepIso() error {
// 	cmd := exec.Command("sudo", "umount", "/var/vcap/cache")
// 	if out, err := cmd.Output(); err != nil {
// 		os.Stdout.Write(out)
// 		return fmt.Errorf("unmounting cf-oss-deps.iso: %s", err)
// 	}
// 	return nil
// }

func (s *start) startGarden(garden client.Client) error {
	// TODO download gdn cli
	// TODO Inform user they need xfsprogs
	var buf bytes.Buffer
	// Add to below? --dns-server=8.8.8.8
	s.gdnServer = exec.Command("sudo", "/var/vcap/gdn-1.12.1", "server", "--bind-socket=/var/vcap/gdn.socket")
	s.gdnServer.Stdout = &buf
	s.gdnServer.Stderr = &buf
	if err := s.gdnServer.Start(); err != nil {
		return fmt.Errorf("starting garden: %s", err)
	}
	fmt.Println("DEBUG: Waiting for Garden")
	if err := gdn.WaitForGarden(garden, 2*time.Second); err != nil {
		s.UI.Say(buf.String())
		s.gdnServer.Process.Kill()
		return fmt.Errorf("starting garden: %s", err)
	}
	if err := ioutil.WriteFile(filepath.Join(s.Config.CFDevHome, "garden.pid"), []byte(fmt.Sprintf("%d", s.gdnServer.Process.Pid)), 0644); err != nil {
		s.gdnServer.Process.Kill()
		return fmt.Errorf("writing garden pid file: %s", err)
	}
	return nil
}
