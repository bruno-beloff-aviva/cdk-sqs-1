package testmessage

import (
	"fmt"
	"time"
)

type TestMessage struct {
	Client  string
	Created string
}

func NewTestMessage(client string) TestMessage {
	time := time.Now().UTC().Format(time.RFC3339)

	return TestMessage{Client: client, Created: time}
}

func (m *TestMessage) String() string {
	return fmt.Sprintf("TestMessage:{Client:%s Created:%s}", m.Client, m.Created)
}
