package syncx

import (
	"errors"
	"sync"
)

var ErrUseOfCleaned = errors.New("using a cleaned resource")

type Resource struct {
	lock    sync.Mutex
	ref     int32
	cleaned bool
	clean   func()
}

func NewResource(clean func()) *Resource {
	return &Resource{
		clean: clean,
	}
}

func (rc *Resource) Use() error {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	if rc.cleaned {
		return ErrUseOfCleaned
	}

	rc.ref++
	return nil
}

func (rc *Resource) Clean() {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	if rc.cleaned {
		return
	}

	rc.ref--
	if rc.ref == 0 {
		rc.cleaned = true
		rc.clean()
	}
}
