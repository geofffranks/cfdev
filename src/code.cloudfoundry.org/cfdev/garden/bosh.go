package garden

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/garden"
)

func DeployBosh(client garden.Client) error {
	// TODO
	//  sudo mount -o loop,ro,uid=dgodd /home/dgodd/.cfdev/cache/cf-oss-deps.iso /var/vcap/cache/

	containerSpec := garden.ContainerSpec{
		Handle:     "deploy-bosh",
		Privileged: true,
		Network:    "10.246.0.0/16",
		Image: garden.ImageRef{
			URI: "/var/vcap/cache/workspace.tar",
		},
		BindMounts: []garden.BindMount{
			{
				SrcPath: "/var/vcap",
				DstPath: "/var/vcap",
				Mode:    garden.BindMountModeRW,
			},
			// TODO macos vs linux and make linux generic to CfdevHome
			// {
			// 	SrcPath: "/var/vcap/cache",
			// 	DstPath: "/var/vcap/cache",
			// 	Mode:    garden.BindMountModeRO,
			// },
			{
				SrcPath: "/home/dgodd/.cfdev/cache",
				DstPath: "/var/vcap/cfdev_cache",
				Mode:    garden.BindMountModeRO,
			},
		},
	}

	container, err := client.Create(containerSpec)
	if err != nil {
		return err
	}

	// process, err := container.Run(garden.ProcessSpec{
	// 	ID:   "deploy-bosh",
	// 	Path: "/bin/mount",
	// 	Args: []string{"-o", "loop,ro", "/var/vcap/cache/cf-oss-deps.iso", "/var/vcap/cache"},
	// 	User: "root",
	// }, garden.ProcessIO{
	// 	Stdout: os.Stdout,
	// 	Stderr: os.Stderr,
	// })
	// if err != nil {
	// 	return err
	// }
	// exitCode, err := process.Wait()
	// if err != nil {
	// 	return err
	// }
	// if exitCode != 0 {
	// 	return fmt.Errorf("process exited with status %v", exitCode)
	// }

	// TODO place socat in workspace.tar
	// curl -L -o /var/vcap/socat https://github.com/andrew-d/static-binaries/raw/master/binaries/linux/x86_64/socat
	// chmod +x /var/vcap/socat

	// _, err = container.Run(garden.ProcessSpec{
	// 	ID:   "deploy-bosh",
	// 	Path: "/var/vcap/socat",
	// 	Args: []string{"tcp-listen:7777,reuseaddr,fork", "unix-connect:/var/vcap/gdn.socket"},
	// 	User: "root",
	// }, garden.ProcessIO{
	// 	Stdout: os.Stdout,
	// 	Stderr: os.Stderr,
	// })
	// if err != nil {
	// 	return err
	// }
	// defer socatProcess.Signal(garden.SignalKill)

	// RUN
	// - ifconfig lo:0 10.0.0.10
	// - /var/vcap/socat tcp-listen:7777,reuseaddr,fork unix-connect:/var/vcap/gdn.socket

	process, err := container.Run(garden.ProcessSpec{
		ID:   "deploy-bosh",
		Path: "/var/vcap/deploy-bosh", // TODO copy back to workspace.tar // "/usr/bin/deploy-bosh",
		User: "root",
	}, garden.ProcessIO{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}
	exitCode, err := process.Wait()
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("process exited with status %v", exitCode)
	}

	client.Destroy("deploy-bosh")

	return nil
}
