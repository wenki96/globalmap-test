package etcd

import "errors"

var (
	PrefixKey  = "/key/"
	PrefixLock = "/lock-key/"
)

var (
	PrefixMapKey    = "/globalmap/"
	PrefixMapLock   = "/lock-globalmap/"
	PrefixResetLock = "/reset-test/"
)

var RefusedByLockError = errors.New("refused by lock")
