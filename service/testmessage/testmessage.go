package testmessage

import (
	"fmt"
	"time"
)

// --------------------------------------------------------------------------------------------------------------------

type TestMessage struct {
	Sent   string
	Path   string
	Client string
}

func NewTestMessage(client string, path string) TestMessage {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	return TestMessage{Sent: now, Path: path, Client: client}
}

func (m *TestMessage) String() string {
	return fmt.Sprintf("TestMessage:{Sent:%s Path:%s Client:%s}", m.Sent, m.Path, m.Client)
}

// --------------------------------------------------------------------------------------------------------------------

type TestReception struct {
	PK         string
	Received   string
	Subscriber string
	TestMessage
}

func GetDeletionKeys() []string {
	return []string{"PK", "Path"}
}

func NewTestReception(subscriber string, message TestMessage) TestReception {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pk := message.Sent + "/" + subscriber

	return TestReception{PK: pk, Received: now, Subscriber: subscriber, TestMessage: message}
}

func (r *TestReception) String() string {
	return fmt.Sprintf("TestReception:{Received:%s Subscriber:%s Sent:%s Path:%s Client:%s}", r.Received, r.Subscriber, r.Sent, r.Path, r.Client)
}

func (r *TestReception) GetKey() map[string]any {
	return map[string]any{"PK": r.PK}
}
