package traffic

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/vsaien/cuter/lib/logx"
)

const httpTimeout = time.Second * 5

var ErrWriteFailed = errors.New("submit failed")

type RemoteWriter struct {
	endpoint string
}

func NewRemoteWriter(endpoint string) Writer {
	return &RemoteWriter{
		endpoint: endpoint,
	}
}

func (rw *RemoteWriter) Write(report *Report) error {
	bs, err := json.Marshal(report)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: httpTimeout,
	}
	resp, err := client.Post(rw.endpoint, "application/json", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logx.Errorf("write report failed, code: %d, reason: %s", resp.StatusCode, resp.Status)
		return ErrWriteFailed
	}

	return nil
}
