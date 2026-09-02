package utils

import (
	"context"
	"errors"
	"fmt"
)

func GetFromContext[T any](ctx context.Context, key string) (*T, error) {
	if v := ctx.Value(key); v != nil {
		val, ok := v.(*T)
		if !ok {
			return nil, fmt.Errorf("context value of wrong type; expected %T, got %T", new(T), v)
		}
		return val, nil
	}
	return nil, errors.New("key not found in context: " + key)
}
