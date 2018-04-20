package garden

import (
	"code.cloudfoundry.org/garden/client/connection"
)

func newGardenConnection() connection.Connection {
	return connection.New("tcp", "localhost:8888")
}
