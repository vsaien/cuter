package service

import (
	"github.com/vsaien/cuter/lib/logx"
	"github.com/vsaien/cuter/lib/prometheus"
	"github.com/vsaien/cuter/lib/traffic"
)

type ServiceConf struct {
	Name       string
	Log        logx.Config
	MetricsUrl string            `json:",optional"`
	Prometheus prometheus.Config `json:",optional"`
}

func (sc ServiceConf) SetUp() error {
	if len(sc.Log.ServiceName) == 0 {
		sc.Log.ServiceName = sc.Name
	}
	if err := logx.SetUp(sc.Log); err != nil {
		return err
	}

	prometheus.StartAgent(sc.Prometheus)
	if len(sc.MetricsUrl) > 0 {
		traffic.SetReportWriter(traffic.NewRemoteWriter(sc.MetricsUrl))
	}

	return nil
}
