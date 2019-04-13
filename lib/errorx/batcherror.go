package errorx

import "bytes"

type BatchError []error

func (be BatchError) Err() error {
	if len(be) > 0 {
		return be
	} else {
		return nil
	}
}

func (be BatchError) Error() string {
	var buf bytes.Buffer

	for i := range be {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(be[i].Error())
	}

	return buf.String()
}
