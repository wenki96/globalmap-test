package logic

import (
	"context"
	"errors"
	"log"

	"main/etcd"
	"main/test/internal/svc"
	"main/test/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type TestSendTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestSendTxLogic {
	return TestSendTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestSendTxLogic) TestSendTx(req types.RequestSendTx) (resp *types.ResponseSendTx, err error) {
	// init distributed lock in etcd
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	s, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		errInfo := err.Error()
		logx.Error(errInfo)
		return &types.ResponseSendTx{}, err
	}
	defer s.Close()
	ctx := context.Background()

	if etcd.ResetKeyUsed(cli) {
		errInfo := "[sendtx] Globalmap is resetting"
		logx.Error(errInfo)
		return &types.ResponseSendTx{}, errors.New(errInfo)
	}

	uuid := uuid.NewString()

	keyLock := etcd.PrefixResetLock + uuid
	mtx := concurrency.NewMutex(s, keyLock)

	// acquire lock
	if err := mtx.Lock(ctx); err != nil {
		errInfo := err.Error()
		logx.Error(errInfo)
		return &types.ResponseSendTx{}, err
	}
	defer mtx.Unlock(ctx)

	return &types.ResponseSendTx{
		Num: 100,
	}, nil
}
