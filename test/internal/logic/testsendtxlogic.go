package logic

import (
	"context"
	"errors"
	"log"
	"time"
	"strconv"
	"fmt"

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

	start1 := time.Now()

	// acquire lock
	if err := mtx.Lock(ctx); err != nil {
		errInfo := err.Error()
		logx.Error(errInfo)
		return &types.ResponseSendTx{}, err
	}
	
	lockArray := make([]*concurrency.Mutex, 0)

	var Users []string
	nums := generateRandomNumber(1, 200000, 3)
	for _, n := range nums {
		Users = append(Users, strconv.Itoa(n))
		// fmt.Println(n)
	}
	//Users := []string{"1", "2", "3"}

	for i := 0; i < 3; i++ {
		keyLock := etcd.PrefixMapLock + Users[i]
		lockAsset := concurrency.NewMutex(s, keyLock)
		if err := lockAsset.TryLock(ctx); err != nil {
			for _, lock := range lockArray {
				lock.Unlock(ctx)
			}
			logx.Error("refused by lock")
			return &types.ResponseSendTx{}, err
		}
		lockArray = append([]*concurrency.Mutex{lockAsset}, lockArray...)
	}

	// time.Sleep(400 * time.Millisecond)

	for _, lock := range lockArray {
		lock.Unlock(ctx)
	}

	mtx.Unlock(ctx)

	end1 := time.Since(start1)
	fmt.Printf("sendtx time %v\n", end1)

	// time.Sleep(200 * time.Millisecond)

	return &types.ResponseSendTx{
		Num: 100,
	}, nil
}
