package garden

import (
	"time"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
)

func NewClient() client.Client {
	return client.New(newGardenConnection())
}

func WaitForGarden(gClient garden.Client) {
	for {
		if err := gClient.Ping(); err == nil {
			return
		}

		time.Sleep(time.Second)
	}
}
