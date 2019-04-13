package service

import "github.com/vsaien/cuter/lib/messages"

type ContentConsumer interface {
	Consume(*messages.JsonMessage) error
}
