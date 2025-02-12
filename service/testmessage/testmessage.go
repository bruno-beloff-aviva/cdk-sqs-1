package testmessage

import (
	"fmt"
	"time"
)

type TestMessage struct {
	Created string
	Path    string
	Client  string
}

func NewTestMessage(client string, path string) TestMessage {
	time := time.Now().UTC().Format(time.RFC3339Nano)

	return TestMessage{Created: time, Path: path, Client: client}
}

func (m *TestMessage) String() string {
	return fmt.Sprintf("TestMessage:{Created:%s Path:%s Client:%s}", m.Created, m.Path, m.Client)
}

type TestReception struct {
	Received string
	TestMessage
}

func NewTestReception(message TestMessage) TestReception {
	time := time.Now().UTC().Format(time.RFC3339Nano)

	return TestReception{Received: time, TestMessage: message}
}

func (m *TestReception) String() string {
	return fmt.Sprintf("TestReception:{Received:%s Created:%s Path:%s Client:%s}", m.Received, m.Created, m.Path, m.Client)
}
