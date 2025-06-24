package repository

import (
	"context"
)

type CtxKey struct{}

type TxManager interface {
	Do(context.Context, func(context.Context) error) error
}

type Transaction interface {
	Commit(context.Context) error
	Rollback(context.Context) error
	Begin() interface{}
}

type CtxManager interface {
	Default(context.Context) Transaction
	ByKey(context.Context, CtxKey) Transaction
	CtxKey() CtxKey
}
