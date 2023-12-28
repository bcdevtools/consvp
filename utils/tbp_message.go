package utils

import (
	"fmt"
	"sync"
	"time"
)

// IToBePrintedMessagesHelper is a helper to queue messages and will print them out when requested by PrintQueuedMessages().
type IToBePrintedMessagesHelper interface {
	AddMessage(message string, error bool)
	PrintQueuedMessages()
}

// TbpMessages implements IToBePrintedMessagesHelper, to be used to queue messages during T-UI active, then print them out when T-UI is closed.
// This is to work around the issue that T-UI will clear the screen and error messages will be lost.
var TbpMessages IToBePrintedMessagesHelper = &toBePrintedMessagesHelper{
	mutex:          &sync.Mutex{},
	queuedMessages: nil,
}

type queuedMessage struct {
	time    time.Time
	message string
	error   bool
}

type toBePrintedMessagesHelper struct {
	mutex          *sync.Mutex
	queuedMessages []*queuedMessage
}

func (t *toBePrintedMessagesHelper) AddMessage(message string, error bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.queuedMessages = append(t.queuedMessages, &queuedMessage{
		time:    time.Now(),
		message: message,
		error:   error,
	})
}

func (t *toBePrintedMessagesHelper) PrintQueuedMessages() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, queuedMessage := range t.queuedMessages {
		msg := fmt.Sprintf("%s: %s", queuedMessage.time.Format("15:04:05"), queuedMessage.message)
		if queuedMessage.error {
			PrintlnStdErr(msg)
		} else {
			fmt.Println(msg)
		}
	}

	t.queuedMessages = nil
}
