package logic

import (
	"github.com/vsaien/cuter/application/test/demo/model"
	"github.com/vsaien/cuter/application/shared/rpcclient/test/demo"
)

type (
	DemoLogic struct {
		demoModel     *model.DemoModel
		demoRpcClient *demo.DemoRpcClient
	}
	DemoRequest struct {
		Token string `path:"token"`
	}
)

func NewDemoLogic(demoModel *model.DemoModel, demoRpcModel *demo.DemoRpcClient) *DemoLogic {
	return &DemoLogic{
		demoModel:     demoModel,
		demoRpcClient: demoRpcModel,
	}
}
func (l *DemoLogic) Demo(r *DemoRequest) (string, error) {
	res, err := l.demoRpcClient.DemoFunc(r.Token)
	if nil != err {
		return "", err
	}
	return res.Token, nil
}
func (l *DemoLogic) DemoTwo(r *DemoRequest) (string, error) {
	return r.Token, nil
}
