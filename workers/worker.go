package workers

import ()

const (
	MAX_RECEIVE_MESSAGE_NUMBER = 10
	VISIBILITY_TIMEOUT_SEC     = 60
	WAIT_TIME_SECONDS          = 20
)

type Worker interface {

	// Start worker
	Start()

	// Stop worker
	Stop()
}
