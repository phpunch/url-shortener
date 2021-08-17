package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
	"url-shortener/model"
)

// Repository is an interface for key-value database
type Repository interface {
	Set(context.Context, *model.UrlObject) (string, error)
	Get(context.Context, string) (*model.UrlObject, error)
	Exists(context.Context, string) (bool, error)
	SAdd(context.Context, string, string) (bool, error)
	SMembers(context.Context, string) ([]string, error)
}

// redisRepository is a storange management
type redisRepository struct {
	Pool *redis.Pool
}

// NewPool initiates a redis pool and its configuration
func NewPool(address string) (Repository, error) {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	return &redisRepository{Pool: pool}, nil
}

func (r *redisRepository) Set(ctx context.Context, o *model.UrlObject) (string, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return "", fmt.Errorf("context expired. err: %v", err)
	}

	jsonBytes, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json, err: %v", err)
	}

	_, err = conn.Do("SET", o.ShortCode, jsonBytes)
	if err != nil {
		return "", fmt.Errorf("failed to set data: %v", err)
	}

	return "", nil
}
func (r *redisRepository) Get(ctx context.Context, key string) (*model.UrlObject, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("context expired. err: %v", err)
	}

	jsonBytes, err := redis.Bytes(conn.Do("SET", key))
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %v", err)
	}

	var o model.UrlObject
	if err = json.Unmarshal(jsonBytes, &o); err != nil {
		return nil, err
	}

	return &o, nil
}
func (r *redisRepository) Exists(ctx context.Context, key string) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}

	value, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, fmt.Errorf("failed to check existance: %v", err)
	}

	return value, nil
}

func (r *redisRepository) SAdd(ctx context.Context, setGroup string, member string) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}

	num, err := redis.Int(conn.Do("SADD", setGroup, member))
	if num != 1 {
		return false, fmt.Errorf("failed to add member, setGroup: %s, member: %s", setGroup, member)
	}
	if err != nil {
		return false, fmt.Errorf("failed to add member, err: %v", err)
	}

	return true, nil
}

func (r *redisRepository) SMembers(ctx context.Context, key string) ([]string, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("context expired. err: %v", err)
	}

	members, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %v", err)
	}

	return members, nil
}
