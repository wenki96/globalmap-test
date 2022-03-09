package logic

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"main/etcd"
	"main/test/internal/svc"
	"main/test/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type TestMultiUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestMultiUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestMultiUserLogic {
	return TestMultiUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

//生成count个[start,end)结束的不重复的随机数
func generateRandomNumber(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}

	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn((end - start)) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}

func (l *TestMultiUserLogic) TestMultiUser(req types.RequestMultiUser) (resp *types.ResponseMultiUser, err error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}})
	if err != nil {
		return &types.ResponseMultiUser{}, err
	}
	defer cli.Close()
	s, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		return &types.ResponseMultiUser{}, err
	}
	defer s.Close()
	ctx := context.Background()

	lockArray := make([]*concurrency.Mutex, 0)

	var Users []string
	nums := generateRandomNumber(1, 100, 2)
	for _, n := range nums {
		Users = append(Users, strconv.Itoa(n))
		// fmt.Println(n)
	}
	//Users := []string{"1", "2", "3"}

	for i := 0; i < 2; i++ {
		keyLock := etcd.PrefixMapLock + Users[i]
		lockAsset := concurrency.NewMutex(s, keyLock)
		if err := lockAsset.TryLock(ctx); err != nil {
			for _, lock := range lockArray {
				lock.Unlock(ctx)
			}
			logx.Error("refused by lock")
			return &types.ResponseMultiUser{}, err
		}
		lockArray = append([]*concurrency.Mutex{lockAsset}, lockArray...)
	}

	time.Sleep(100 * time.Millisecond)

	for _, lock := range lockArray {
		lock.Unlock(ctx)
	}

	return &types.ResponseMultiUser{
		Num: 1,
	}, nil
}
