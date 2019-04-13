package service

import (
	"log"

	"github.com/vsaien/cuter/lib/system"
	"github.com/vsaien/cuter/lib/threading"
)

type (
	Starter interface {
		Start()
	}

	Stopper interface {
		Stop()
	}

	Service interface {
		Starter
		Stopper
	}

	ServiceGroup struct {
		services []Service
	}
)

func NewServiceGroup() *ServiceGroup {
	return &ServiceGroup{}
}

func (sg *ServiceGroup) Add(service Service) {
	sg.services = append(sg.services, service)
}

// There should not be any logic code after calling this method, because this method is a blocking one.
// Also, quitting this method will close the logx output.
func (sg *ServiceGroup) Start() {
	system.AddShutdownListener(func() {
		log.Println("Shutting down...")
		sg.Stop()
	})

	sg.doStart()
}

func (sg *ServiceGroup) Stop() {
	for _, service := range sg.services {
		service.Stop()
	}
}

func (sg *ServiceGroup) doStart() {
	routineGroup := threading.NewRoutineGroup()

	for i := range sg.services {
		service := sg.services[i]
		routineGroup.RunSafe(func() {
			service.Start()
		})
	}

	routineGroup.WaitForDone()
}

func WithStart(start func()) Service {
	return startOnlyService{
		start: start,
	}
}

func WithStarter(start Starter) Service {
	return starterOnlyService{
		Starter: start,
	}
}

type (
	stopper struct {
	}

	startOnlyService struct {
		start func()
		stopper
	}

	starterOnlyService struct {
		Starter
		stopper
	}
)

func (s stopper) Stop() {
}

func (s startOnlyService) Start() {
	s.start()
}
