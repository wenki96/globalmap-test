package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"main/etcd"
	"main/test/internal/svc"
	"main/test/internal/types"

	"github.com/google/uuid"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/zeromicro/go-zero/core/logx"
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

// //获取redis连接池
// func newPool() *redis.Pool {
// 	return &redis.Pool{
// 		MaxIdle:     3,
// 		IdleTimeout: time.Duration(24) * time.Second,
// 		Dial: func() (redis.Conn, error) {
// 			c, err := redis.Dial("tcp", "localhost:6379")
// 			if err != nil {
// 				panic(err.Error())
// 				return nil, err
// 			}
// 			return c, err
// 		},
// 		TestOnBorrow: func(c redis.Conn, t time.Time) error {
// 			_, err := c.Do("PING")
// 			if err != nil {
// 				return err
// 			}
// 			return err
// 		},
// 	}
// }

func (l *TestSendTxLogic) TestSendTx(req types.RequestSendTx) (resp *types.ResponseSendTx, err error) {
	start1 := time.Now()

	ctx := context.Background()

	if etcd.G_etcd.ResetKeyUsed() {
		errInfo := "[sendtx] Globalmap is resetting"
		logx.Error(errInfo)
		return &types.ResponseSendTx{}, errors.New(errInfo)
	}

	uuid := uuid.NewString()

	keyLock := etcd.PrefixResetLock + uuid
	mtx := concurrency.NewMutex(etcd.G_etcd.S, keyLock)

	// pool := newPool() // or, pool := redigo.NewPool(...)

	// rs := redsync.New([]redsync.Pool{pool})

	// // Create an instance of redisync to be used to obtain a mutual exclusion
	// // lock.

	// mutexname := "my-global-mutex"
	// mutex := rs.NewMutex(mutexname)

	// if err := mutex.Lock(); err != nil {
	// 	panic(err)
	// }
	// if ok, err := mutex.Unlock(); !ok || err != nil {
	// 	panic("unlock failed")
	// }

	// var putResp *clientv3.PutResponse
	// // 实例化一个用于操作ETCD的KV
	// kv := clientv3.NewKV(cli)

	// _, err = kv.Put(ctx, "/test/", "1")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return &types.ResponseSendTx{}, errors.New(err.Error())
	// }

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
		lockAsset := concurrency.NewMutex(etcd.G_etcd.S, keyLock)
		if err := lockAsset.TryLock(ctx); err != nil {
			for _, lock := range lockArray {
				lock.Unlock(ctx)
			}
			logx.Error("refused by lock")
			return &types.ResponseSendTx{}, err
		}
		lockArray = append([]*concurrency.Mutex{lockAsset}, lockArray...)
	}

	for _, lock := range lockArray {
		lock.Unlock(ctx)
	}

	mtx.Unlock(ctx)

	// if ok, err := mutex.Unlock(); !ok || err != nil {
	// 	panic("unlock failed")
	// }

	end1 := time.Since(start1)
	fmt.Printf("sendtx time %v\n", end1)

	// time.Sleep(200 * time.Millisecond)

	return &types.ResponseSendTx{
		Num: 100,
	}, nil
}
