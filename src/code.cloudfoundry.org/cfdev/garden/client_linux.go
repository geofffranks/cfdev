package garden

import (
	"code.cloudfoundry.org/garden/client/connection"
)

func newGardenConnection() connection.Connection {
	return connection.New("unix", "/var/vcap/gdn.socket")
}
