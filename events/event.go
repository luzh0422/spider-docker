package events

import ()

type (
	Event struct {
		EventId int
	}
	Container struct {
		CrawlId string
	}
)

const (
	EVENT_BASE = 0x10000000

	/**
	 * run docker
	 */
	DOCKER_RUN = EVENT_BASE + 1002
	/**
	 * delete container
	 */
	DELETE_CONTAINER = EVENT_BASE + 1003
)
