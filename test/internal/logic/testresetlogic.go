package logic

import (
	"context"
	"time"

	"main/etcd"
	"main/test/internal/svc"
	"main/test/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestResetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestResetLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestResetLogic {
	return TestResetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestResetLogic) TestReset(req types.RequestReset) (resp *types.ResponseReset, err error) {
	for {
		err = etcd.G_etcd.ResetGlobalMap()
		if err != nil {
			logx.Error("what the fuck")
		}
		time.Sleep(2 * time.Second)
	}
}
