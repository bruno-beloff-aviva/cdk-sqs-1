package testmessage

import (
	"fmt"
	"time"
)

type TestMessage struct {
	Client  string
	Path    string
	Created string
}

func NewTestMessage(client string, path string) TestMessage {
	time := time.Now().UTC().Format(time.RFC3339)

	return TestMessage{Client: client, Path: path, Created: time}
}

func (m *TestMessage) String() string {
	return fmt.Sprintf("TestMessage:{Client:%s Path:%s Created:%s}", m.Client, m.Path, m.Created)
}
