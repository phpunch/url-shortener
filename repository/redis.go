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
	Set(ctx context.Context, key string, o interface{}, expiry *time.Time) (bool, error)
	Get(context.Context, string) (*model.UrlObject, error)
	Del(context.Context, string) (bool, error)
	Exists(context.Context, string) (bool, error)
	// SAdd(context.Context, string, string) (bool, error)
	// SIsMember(ctx context.Context, key string, member string) (bool, error)
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

func (r *redisRepository) Set(ctx context.Context, key string, o interface{}, expiry *time.Time) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	jsonBytes, err := json.Marshal(o)
	if err != nil {
		return false, fmt.Errorf("failed to marshal json, err: %v", err)
	}

	// set url object
	_, err = conn.Do("SET", key, jsonBytes)
	if err != nil {
		return false, fmt.Errorf("failed to set data: %v", err)
	}

	// set expiry only specified
	var defaultTime time.Time
	if expiry != nil || expiry.Unix() != defaultTime.Unix() {
		_, err = conn.Do("EXPIREAT", key, expiry.Unix())
		if err != nil {
			return false, fmt.Errorf("failed to set expire at: %v", err)
		}
	}

	return true, nil
}
func (r *redisRepository) Get(ctx context.Context, key string) (*model.UrlObject, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	jsonBytes, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %v", err)
	}

	var o model.UrlObject
	if err = json.Unmarshal(jsonBytes, &o); err != nil {
		return nil, err
	}

	return &o, nil
}
func (r *redisRepository) Del(ctx context.Context, key string) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	deleteKeys, err := redis.Int(conn.Do("DEL", key))
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
// 	defer conn.Close()

// 	num, err := redis.Int(conn.Do("SADD", setGroup, member))
// 	if num != 1 {
// 		return false, fmt.Errorf("failed to add member, setGroup: %s, member: %s", setGroup, member)
// 	}
// 	if err != nil {
// 		return false, fmt.Errorf("failed to add member, err: %v", err)
// 	}

// 	return true, nil
// }

// func (r *redisRepository) SIsMember(ctx context.Context, key string, member string) (bool, error) {
// 	conn, err := r.Pool.GetContext(ctx)
// 	if err != nil {
// 		return false, fmt.Errorf("context expired. err: %v", err)
// 	}
// 	defer conn.Close()

// 	members, err := redis.Int(conn.Do("SMEMBERS", key))
// 	if err != nil {
// 		return false, fmt.Errorf("failed to search members in set: %v", err)
// 	}
// 	if members != 1 {
// 		return false, fmt.Errorf("%s is not in %s", member, key)
// 	}

// 	return true, nil
// }
func (r *redisRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %v", err)
	}

	return keys, nil
}
