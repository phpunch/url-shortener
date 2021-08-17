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
	Set(context.Context, string, string, *model.UrlObject) (string, error)
	Get(context.Context, string, string) (*model.UrlObject, error)
	Del(context.Context, string, string) (bool, error)
	Exists(context.Context, string) (bool, error)
	// SAdd(context.Context, string, string) (bool, error)
	// SMembers(context.Context, string) ([]string, error)
	Keys(context.Context, string) ([]string, error)
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

func (r *redisRepository) Set(ctx context.Context, prefix string, key string, o *model.UrlObject) (string, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return "", fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	jsonBytes, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json, err: %v", err)
	}

	// set url object
	_, err = conn.Do("SET", prefix+key, jsonBytes)
	if err != nil {
		return "", fmt.Errorf("failed to set data: %v", err)
	}

	// set expiry only specified
	var defaultTime time.Time
	if o.Expiry.Unix() != defaultTime.Unix() {
		_, err = conn.Do("EXPIREAT", prefix+key, o.Expiry.Unix())
		if err != nil {
			return "", fmt.Errorf("failed to set expire at: %v", err)
		}
	}

	return "", nil
}
func (r *redisRepository) Get(ctx context.Context, prefix string, key string) (*model.UrlObject, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	jsonBytes, err := redis.Bytes(conn.Do("GET", prefix+key))
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %v", err)
	}

	var o model.UrlObject
	if err = json.Unmarshal(jsonBytes, &o); err != nil {
		return nil, err
	}

	return &o, nil
}
func (r *redisRepository) Del(ctx context.Context, prefix string, key string) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	deleteKeys, err := redis.Int(conn.Do("DEL", prefix+key))
	if deleteKeys != 1 {
		return false, fmt.Errorf("failed to delete key. err: key not found")
	}
	if err != nil {
		return false, fmt.Errorf("failed to delete key. err: %v", err)
	}

	return true, nil
}
func (r *redisRepository) Exists(ctx context.Context, key string) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	value, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, fmt.Errorf("failed to check existance: %v", err)
	}

	return value, nil
}

// func (r *redisRepository) SAdd(ctx context.Context, setGroup string, member string) (bool, error) {
// 	conn, err := r.Pool.GetContext(ctx)
// 	if err != nil {
// 		return false, fmt.Errorf("context expired. err: %v", err)
// 	}

// 	num, err := redis.Int(conn.Do("SADD", setGroup, member))
// 	if num != 1 {
// 		return false, fmt.Errorf("failed to add member, setGroup: %s, member: %s", setGroup, member)
// 	}
// 	if err != nil {
// 		return false, fmt.Errorf("failed to add member, err: %v", err)
// 	}

// 	return true, nil
// }

// func (r *redisRepository) SMembers(ctx context.Context, key string) ([]string, error) {
// 	conn, err := r.Pool.GetContext(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("context expired. err: %v", err)
// 	}

// 	members, err := redis.Strings(conn.Do("SMEMBERS", key))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get members: %v", err)
// 	}

// 	return members, nil
// }
func (r *redisRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("context expired. err: %v", err)
	}

	keys, err := redis.Strings(conn.Do("KEYS", pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %v", err)
	}

	return keys, nil
}
