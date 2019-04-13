package service

import (
	"fmt"
	"strings"

	"github.com/vsaien/cuter/lib/messages"
	"github.com/vsaien/cuter/lib/queue"
)

type ConsumerRouter struct {
	routes          map[string]ContentConsumer
	defaultConsumer ContentConsumer
}

func NewConsumerRouter() *ConsumerRouter {
	return &ConsumerRouter{
		routes: make(map[string]ContentConsumer),
	}
}

func (router *ConsumerRouter) Add(name string, consumer ContentConsumer) {
	router.routes[strings.ToUpper(name)] = consumer
}

func (router *ConsumerRouter) Consume(str string) error {
	if jsonMessage, err := messages.NewJsonMessage(str); err != nil {
		return err
	} else {
		return router.route(jsonMessage)
	}
}

func (router *ConsumerRouter) Factory() queue.ConsumerFactory {
	return func() (queue.Consumer, error) {
		return router, nil
	}
}

func (router *ConsumerRouter) OnEvent(event interface{}) {
}

func (router *ConsumerRouter) SetDefaultConsumer(consumer ContentConsumer) {
	router.defaultConsumer = consumer
}

func (router *ConsumerRouter) route(message *messages.JsonMessage) error {
	tagName := strings.ToUpper(message.TagName)
	if consumer, ok := router.routes[tagName]; ok {
		return consumer.Consume(message)
	} else if router.defaultConsumer != nil {
		return router.defaultConsumer.Consume(message)
	}

	return RouteError{tagName}
}

type RouteError struct {
	TagName string
}

func (e RouteError) Error() string {
	return fmt.Sprintf("No route defined for %s", e.TagName)
}
