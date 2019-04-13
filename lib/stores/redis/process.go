package redis

import (
	"fmt"
	"strings"
	"time"

	"github.com/vsaien/cuter/lib/logx"

	red "github.com/go-redis/redis"
)

func process(oldProcess func(red.Cmder) error) func(red.Cmder) error {
	return func(cmd red.Cmder) error {
		start := time.Now()

		defer func() {
			duration := time.Since(start)
			if duration > slowThreshold {
				var buf strings.Builder
				buf.WriteString(cmd.Name())
				for _, arg := range cmd.Args() {
					buf.WriteString(fmt.Sprintf(" %v", arg))
				}
				logx.Slowf("[REDIS] slowcall(%s) on executing: %s", duration, buf.String())
			}
		}()

		return oldProcess(cmd)
	}
}
