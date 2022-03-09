package etcd

import (
	"context"
	"fmt"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func ResetKeyUsed(cli *clientv3.Client) bool {
	var getResp *clientv3.GetResponse
	// 实例化一个用于操作ETCD的KV
	kv := clientv3.NewKV(cli)

	getResp, err := kv.Get(context.TODO(), PrefixResetLock)
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

	ctx := context.Background()

	keyLock := PrefixLock + key
	l := concurrency.NewMutex(s, keyLock)

	// acquire lock (or wait to have it)
	if err := l.Lock(ctx); err != nil {
		return "", err
	}

	fmt.Println("[Get] acquired lock for get ", keyLock)

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

	if err := l.Unlock(ctx); err != nil {
		return "", err
	}

	fmt.Println("[Get] released lock for get ", keyLock)
	fmt.Println()

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

	ctx := context.Background()

	keyLock := PrefixLock + key
	l := concurrency.NewMutex(s, keyLock)

	// acquire lock (or wait to have it)
	if err := l.Lock(ctx); err != nil {
		return err
	}

	fmt.Println("[Update] acquired lock for update ", keyLock)

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

	if err := l.Unlock(ctx); err != nil {
		return err
	}

	fmt.Println("[Update] released lock for update ", keyLock)
	fmt.Println()

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

	ctx := context.Background()

	keyLock := PrefixLock + key
	l := concurrency.NewMutex(s, keyLock)

	// acquire lock (or wait to have it)
	if err := l.Lock(ctx); err != nil {
		return err
	}

	fmt.Println("[Delete] acquired lock for delete ", keyLock)

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

	if err := l.Unlock(ctx); err != nil {
		return err
	}

	fmt.Println("[Delete] released lock for delete ", keyLock)
	fmt.Println()

	return nil
}

func setResetKey(cli *clientv3.Client, key string) {
	kv := clientv3.NewKV(cli)
	if _, err := kv.Put(context.TODO(), PrefixResetLock, key, clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
		return
	}
}

func ResetGlobalMap(prefixKeyLock, prefixKey string) (err error) {
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

	gresp, err := cli.Get(context.TODO(), PrefixKey, clientv3.WithPrefix())
	if err != nil {
		return err
	} else {
		for _, kv := range gresp.Kvs {
			mtx := concurrency.NewMutex(s, string(kv.Key))

			// acquire lock
			if err := mtx.Lock(ctx); err != nil {
				return err
			}

			lockArray = append([]*concurrency.Mutex{mtx}, lockArray...)
		}
	}

	fmt.Println("[Reset] acquired lock for reset ", prefixKeyLock)

	kv := clientv3.NewKV(cli)
	res, err := kv.Delete(context.TODO(), prefixKey, clientv3.WithPrevKV(), clientv3.WithPrefix())
	if err != nil {
		return err
	} else {
		fmt.Printf("[Reset] delete %d keys\n", res.Deleted)
		for _, preKv := range res.PrevKvs {
			fmt.Printf("[Reset] del key: %s, value: %s\n", preKv.Key, preKv.Value)
		}
	}

	for _, mtx := range lockArray {
		mtx.Unlock(ctx)
	}

	setResetKey(cli, "0")

	fmt.Println("[Reset] released lock for reset ", prefixKeyLock)
	fmt.Println()

	return nil
}
