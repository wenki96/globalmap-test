package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Etcd struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	S      *concurrency.Session
}

// 全局单例
var (
	G_etcd *Etcd
)

// 初始化
func InitEtcd() (err error) {
	// 初始化配置
	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379"},          // 集群地址
		DialTimeout: time.Duration(5) * time.Millisecond, // 连接超时
	}

	// 建立连接
	client, err := clientv3.New(config)
	if err != nil {
		fmt.Println(err)
	}

	// 得到KV和Lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	session, err := concurrency.NewSession(client, concurrency.WithTTL(10))

	// 赋值单例
	G_etcd = &Etcd{
		client: client,
		kv:     kv,
		lease:  lease,
		S:      session,
	}
	return
}

func (etcd *Etcd) ResetKeyUsed() bool {
	getResp, err := etcd.kv.Get(context.TODO(), PrefixLock)
	if err != nil {
		fmt.Println(err)
		return true
	}

	// 输出本次的Revision
	if getResp.Kvs != nil {
		// fmt.Println(getResp.Kvs[0].Value)
		return string(getResp.Kvs[0].Value) == "1"
	}

	return false
}

func (etcd *Etcd) GetGlobalMap(key string) (value string, err error) {
	var getResp *clientv3.GetResponse

	if getResp, err = etcd.kv.Get(context.TODO(), key); err != nil {
		fmt.Println(err)
		return
	}

	// 输出本次的Revision
	if getResp.Kvs != nil {
		fmt.Printf("[Get] Key : %s, Value : %s \n", getResp.Kvs[0].Key, getResp.Kvs[0].Value)
	}
	return string(getResp.Kvs[0].Value), nil
}

func (etcd *Etcd) UpdateGlobalMap(key, value string) (err error) {
	var putResp *clientv3.PutResponse

	if putResp, err = etcd.kv.Put(context.TODO(), key, value, clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(putResp.Header.Revision)
	if putResp.PrevKv != nil {
		fmt.Printf("[Update] preValue: %s CreateRevision : %d  ModRevision: %d  Version: %d \n",
			putResp.PrevKv.Value, putResp.PrevKv.CreateRevision, putResp.PrevKv.ModRevision, putResp.PrevKv.Version)
	}
	fmt.Println("[Update] curValue: ", value)

	return nil
}

func (etcd *Etcd) DeleteGlobalMap(key string) (err error) {
	res, err := etcd.kv.Delete(context.TODO(), key)
	if err != nil {
		return err
	} else {
		fmt.Printf("[Delete] delete %d key\n", res.Deleted)
		for _, preKv := range res.PrevKvs {
			fmt.Printf("[Delete] del key: %s, value: %s\n", preKv.Key, preKv.Value)
		}
	}

	return nil
}

func (etcd *Etcd) setResetKey(key string) {
	if _, err := etcd.kv.Put(context.TODO(), PrefixLock, key); err != nil {
		fmt.Println(err)
		return
	}
}

func (etcd *Etcd) ResetGlobalMap() (err error) {
	ctx := context.Background()

	lockArray := make([]*concurrency.Mutex, 0)

	// acquire lock (or wait to have it)
	etcd.setResetKey("1")

	start1 := time.Now()

	gresp, err := etcd.kv.Get(ctx, PrefixResetLock, clientv3.WithPrefix())
	if err != nil {
		return err
	} else {
		fmt.Printf("waiting for %d locks\n", len(gresp.Kvs))

		cnt := 0

		for _, kv := range gresp.Kvs {
			startget := time.Now()

			keys, _ := etcd.kv.Get(ctx, string(kv.Key))

			endget := time.Since(startget)
			fmt.Printf("get time %v\n", endget)

			if keys.Count == 0 {
				continue
			} else {
				cnt++
				mtx := concurrency.NewMutex(etcd.S, string(kv.Key))

				// acquire lock
				if err := mtx.Lock(ctx); err != nil {
					return err
				}

				lockArray = append([]*concurrency.Mutex{mtx}, lockArray...)
			}

		}

		fmt.Printf("lock nums %d\n", cnt)
	}

	end1 := time.Since(start1)
	fmt.Printf("lock time %v\n", end1)

	res, err := etcd.kv.Delete(context.TODO(), PrefixKey, clientv3.WithPrevKV(), clientv3.WithPrefix())
	if err != nil {
		return err
	} else {
		// fmt.Printf("[Reset] delete %d keys\n", res.Deleted)
		for _, preKv := range res.PrevKvs {
			fmt.Printf("[Reset] del key: %s, value: %s\n", preKv.Key, preKv.Value)
		}
	}

	start2 := time.Now()

	for _, mtx := range lockArray {
		mtx.Unlock(ctx)
	}

	end2 := time.Since(start2)
	end3 := time.Since(start1)

	fmt.Printf("unlock time %v\n", end2)

	fmt.Printf("total time %v\n", end3)
	fmt.Println()

	time.Sleep(1 * time.Second)

	etcd.setResetKey("0")

	return nil
}
