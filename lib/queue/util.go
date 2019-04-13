package queue

import "bytes"

func generateName(pushers []QueuePusher) string {
	var buf bytes.Buffer

	first := true
	for _, pusher := range pushers {
		if first {
			first = false
		} else {
			buf.WriteByte(',')
		}

		buf.WriteString(pusher.Name())
	}

	return buf.String()
}
