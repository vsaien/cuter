package breaker

import (
	"errors"
	"time"
)

type (
	BreakerConfig struct {
		Name        string
		Enable      bool   `json:",default=true"`
		MaxRequests uint32 `json:",default=3"`
		Interval    int    `json:",default=5"`
		Timeout     int    `json:",default=10"`
	}

	Breakers []BreakerConfig
)

func (b Breakers) Setup() error {
	for _, setting := range b {
		if len(setting.Name) == 0 {
			return errors.New("no name specified in breaker setting")
		}

		if !setting.Enable {
			NoBreakFor(setting.Name)
			continue
		}

		SetBreaker(Settings{
			Name:        setting.Name,
			MaxRequests: setting.MaxRequests,
			Interval:    time.Duration(setting.Interval) * time.Second,
			Timeout:     time.Duration(setting.Timeout) * time.Second,
		})
	}

	return nil
}
