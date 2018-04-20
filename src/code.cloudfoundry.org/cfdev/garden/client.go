package garden

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
)

func NewClient() client.Client {
	return client.New(newGardenConnection())
}

func WaitForGarden(gClient garden.Client, timeout time.Duration) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	after := time.After(timeout)
	for {
		select {
		case <-ticker.C:
			if err := gClient.Ping(); err == nil {
				return nil
			}
		case <-after:
			return fmt.Errorf("timedout in %v", timeout)
		}
	}
}
