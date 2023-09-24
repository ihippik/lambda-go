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
	return ":" + strconv.Itoa(m.port)
}

func (m metaData) short() string {
	return m.containerID[:5]
}
