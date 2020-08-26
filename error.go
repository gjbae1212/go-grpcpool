package grpcpool

import "errors"

var (
	ErrorInvalidParams = errors.New("[err][grpcpool] invalid params.")
	ErrorPoolEmpty     = errors.New("[err][grpcpool] pool empty.")
)
