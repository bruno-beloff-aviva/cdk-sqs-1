package testreception

import (
	"fmt"
	"sqstest/service/testmessage"
	"time"
)

var DeletionKeys = []string{"PK", "Received"}

type TestReception struct {
	testmessage.TestMessage
	PK         string
	Received   string
	Subscriber string
}

func NewTestReception(subscriber string, message testmessage.TestMessage) TestReception {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pk := message.Sent + "/" + subscriber

	return TestReception{TestMessage: message, PK: pk, Received: now, Subscriber: subscriber}
}

func (r *TestReception) String() string {
	return fmt.Sprintf("TestReception:{Received:%s Subscriber:%s Sent:%s Path:%s Client:%s}", r.Received, r.Subscriber, r.Sent, r.Path, r.Client)
}

func (r *TestReception) GetPartitionKey() map[string]any {
	return map[string]any{"PK": r.PK}
}
