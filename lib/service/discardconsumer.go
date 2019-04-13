package service

import (
	"log"

	"github.com/vsaien/cuter/lib/messages"
)

var DiscardConsumer = &discardConsumer{}

type discardConsumer struct{}

func (c *discardConsumer) Consume(message *messages.JsonMessage) error {
	log.Printf("Warning: discarding %s\n", message.Raw)
	return nil
}
