package testmessage

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	message := NewTestMessage("client", "path")
	fmt.Println(message.String())

	assert.Equal(t, message.Client, "client")
	assert.Equal(t, message.Path, "path")
}

func TestNewMessageJSON(t *testing.T) {
	var message TestMessage
	var jmsg []byte
	var err error

	message = NewTestMessage("client", "path")
	jmsg, err = json.Marshal(message)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	fmt.Println(strmsg)

	err = json.Unmarshal([]byte(strmsg), &message)
	if err != nil {
		panic(err)
	}

	fmt.Println(message.String())

	assert.Equal(t, message.Client, "client")
	assert.Equal(t, message.Path, "path")
}

func TestNewReception(t *testing.T) {
	message := NewTestMessage("client", "path")
	reception := NewTestReception("sub1", message)

	fmt.Println(reception.String())

	assert.Equal(t, message.Client, "client")
	assert.Equal(t, message.Path, "path")
}

func TestNewReceptionJSON(t *testing.T) {
	var message TestMessage
	var reception TestReception
	var jmsg []byte
	var err error

	message = NewTestMessage("client", "path")
	time.Sleep(1 * time.Second)

	reception = NewTestReception("sub1", message)

	jmsg, err = json.Marshal(reception)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	fmt.Println(strmsg)

	err = json.Unmarshal([]byte(strmsg), &reception)
	if err != nil {
		panic(err)
	}

	fmt.Println(reception.String())

	assert.Equal(t, reception.Client, "client")
	assert.Equal(t, reception.Path, "path")
}
