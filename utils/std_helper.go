package utils

import (
	"fmt"
	"sync"
	"time"
)

// IStdHelper is a helper to print stdout & stderr with ability to enable queue messages then will print them out when requested by PrintQueuedMessages().
type IStdHelper interface {
	// Println prints a message to stdout or queue message if queue is enabled.
	Println(message any)

	// PrintlnStdErr prints a message to stderr or queue message if queue is enabled.
	PrintlnStdErr(message any)

	// EnableQueue enables queue.
	EnableQueue()

	// PrintQueuedMessages prints queued messages.
	PrintQueuedMessages()
}

// StdHelper implements IStdHelper, to be used to queue messages during T-UI active, then print them out when T-UI is closed.
// This is to work around the issue that T-UI will clear the screen and error messages will be lost.
var StdHelper IStdHelper = &stdHelper{
	mutex:          &sync.RWMutex{},
	queuedMessages: nil,
}

type queuedMessage struct {
	time    time.Time
	message any
	error   bool
}

type stdHelper struct {
	mutex          *sync.RWMutex
	enabledQueue   bool
	queuedMessages []*queuedMessage
}

func (h *stdHelper) Println(message any) {
	if !h.isQueueEnabled() {
		fmt.Println(message)
		return
	}

	h.queueMessage(message, false)
}

func (h *stdHelper) PrintlnStdErr(message any) {
	if !h.isQueueEnabled() {
		PrintlnStdErr(message)
		return
	}

	h.queueMessage(message, true)
}

func (h *stdHelper) EnableQueue() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.enabledQueue = true
}

func (h *stdHelper) isQueueEnabled() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.enabledQueue
}

func (h *stdHelper) queueMessage(message any, error bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.queuedMessages = append(h.queuedMessages, &queuedMessage{
		time:    time.Now(),
		message: message,
		error:   error,
	})
}

func (h *stdHelper) PrintQueuedMessages() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, qm := range h.queuedMessages {
		msg := fmt.Sprintf("%s: %s", qm.time.Format("15:04:05"), qm.message)
		if qm.error {
			PrintlnStdErr(msg)
		} else {
			fmt.Println(msg)
		}
	}

	h.queuedMessages = nil
}
