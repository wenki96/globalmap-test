package logic

import (
	"context"
	"time"

	"main/etcd"
	"main/test/internal/svc"
	"main/test/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type TestConstUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestConstUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestConstUserLogic {
	return TestConstUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestConstUserLogic) TestConstUser(req types.RequestConstUser) (resp *types.ResponseConstUser, err error) {

	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}})
	if err != nil {
		return &types.ResponseConstUser{}, err
	}
	defer cli.Close()
	s, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		return &types.ResponseConstUser{}, err
	}
	defer s.Close()
	ctx := context.Background()

	lockArray := make([]*concurrency.Mutex, 0)

	Users := []string{"1", "2", "3"}

	for i := 0; i < 3; i++ {
		keyLock := etcd.PrefixMapLock + Users[i]
		lockAsset := concurrency.NewMutex(s, keyLock)
		if err := lockAsset.TryLock(ctx); err != nil {
			for _, lock := range lockArray {
				lock.Unlock(ctx)
			}
			logx.Error("refused by lock")
			return &types.ResponseConstUser{}, err
		}
		lockArray = append([]*concurrency.Mutex{lockAsset}, lockArray...)
	}

	time.Sleep(200 * time.Millisecond)

	for _, lock := range lockArray {
		lock.Unlock(ctx)
	}

	return &types.ResponseConstUser{
		Num: 1,
	}, nil

	return
}
