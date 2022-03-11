package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func ResetKeyUsed(cli *clientv3.Client) bool {
	var getResp *clientv3.GetResponse
	// 实例化一个用于操作ETCD的KV
	kv := clientv3.NewKV(cli)

	getResp, err := kv.Get(context.TODO(), PrefixLock)
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

func GetGlobalMap(key string) (value string, err error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// create a sessions to aqcuire a lock
	s, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		return "", err
	}
	defer s.Close()

	var getResp *clientv3.GetResponse
	// 实例化一个用于操作ETCD的KV
	kv := clientv3.NewKV(cli)

	if getResp, err = kv.Get(context.TODO(), key); err != nil {
		fmt.Println(err)
		return
	}

	// 输出本次的Revision
	if getResp.Kvs != nil {
		fmt.Printf("[Get] Key : %s, Value : %s \n", getResp.Kvs[0].Key, getResp.Kvs[0].Value)
	}
	return string(getResp.Kvs[0].Value), nil
}

func UpdateGlobalMap(key, value string) (err error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// create a sessions to aqcuire a lock
	s, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		return err
	}
	defer s.Close()

	var putResp *clientv3.PutResponse
	// 实例化一个用于操作ETCD的KV
	kv := clientv3.NewKV(cli)

	if putResp, err = kv.Put(context.TODO(), key, value, clientv3.WithPrevKV()); err != nil {
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

func DeleteGlobalMap(key string) (err error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// create a sessions to aqcuire a lock
	s, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		return err
	}
	defer s.Close()

	// 实例化一个用于操作ETCD的KV
	kv := clientv3.NewKV(cli)

	res, err := kv.Delete(context.TODO(), key)
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

func setResetKey(cli *clientv3.Client, key string) {
	kv := clientv3.NewKV(cli)
	if _, err := kv.Put(context.TODO(), PrefixLock, key, clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
		return
	}
}

func ResetGlobalMap() (err error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	s, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		return err
	}
	defer s.Close()

	ctx := context.Background()

	lockArray := make([]*concurrency.Mutex, 0)

	// acquire lock (or wait to have it)
	setResetKey(cli, "1")

	start1 := time.Now()

	gresp, err := cli.Get(ctx, PrefixResetLock, clientv3.WithPrefix())
	if err != nil {
		return err
	} else {
		fmt.Printf("waiting for %d locks\n", len(gresp.Kvs))

		cnt := 0

		for _, kv := range gresp.Kvs {
			startget := time.Now()

			keys, _ := cli.Get(ctx, string(kv.Key))

			endget := time.Since(startget)
			fmt.Printf("get time %v\n", endget)

			if keys.Count == 0 {
				continue
			} else {
				cnt++
				mtx := concurrency.NewMutex(s, string(kv.Key))

				// acquire lock
				if err := mtx.Lock(ctx); err != nil {
					return err
				}

				lockArray = append([]*concurrency.Mutex{mtx}, lockArray...)
			}

		}

		// fmt.Printf("lock nums %d\n", cnt)
	}

	end1 := time.Since(start1)
	fmt.Printf("lock time %v\n", end1)

	kv := clientv3.NewKV(cli)
	res, err := kv.Delete(context.TODO(), PrefixKey, clientv3.WithPrevKV(), clientv3.WithPrefix())
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

	setResetKey(cli, "0")

	return nil
}
