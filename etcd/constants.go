package etcd

import "errors"

var (
	PrefixKey  = "/key/"
	PrefixLock = "/lock-key/"
)

var (
	PrefixMapKey  = "/globalmap/"
	PrefixMapLock = "/lock-globalmap/"
)

var RefusedByLockError = errors.New("refused by lock")
