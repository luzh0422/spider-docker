package events

import (
	"github.com/spider-docker/models"
)

type (
	SocialEvent struct {
		Event
		CrawlUrl *models.CrawlUrl
	}
	ContainerEvent struct {
		Container
	}
)
