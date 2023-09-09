package lambda

import "strconv"

type metaData struct {
	containerID string
	port        int
}

func newMetaData(containerID string, port int) *metaData {
	return &metaData{containerID: containerID, port: port}
}

func (m metaData) address() string {
	return "http://localhost:" + strconv.Itoa(m.port)
}
