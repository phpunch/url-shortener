package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"reflect"
	"time"
)

// Repository is an interface for key-value database
type Repository interface {
	Set(ctx context.Context, key string, o interface{}, expiry *time.Time) (bool, error)
	Get(ctx context.Context, key string, v interface{}) error
	MGet(ctx context.Context, keys []interface{}, v interface{}) error
	Del(context.Context, string) (bool, error)
	Exists(context.Context, string) (bool, error)
	SAdd(ctx context.Context, key string, member string) (bool, error)
	SIsMember(ctx context.Context, key string, member string) (bool, error)
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

// Set sets an object to a specified key with expiry
func (r *redisRepository) Set(ctx context.Context, key string, object interface{}, expiry *time.Time) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	jsonBytes, err := json.Marshal(object)
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
	if expiry != nil && expiry.Unix() != defaultTime.Unix() {
		_, err = conn.Do("EXPIREAT", key, expiry.Unix())
		if err != nil {
			return false, fmt.Errorf("failed to set expire at: %v", err)
		}
	}

	return true, nil
}

// Get unmarshals a value got from Redis to value `v`
func (r *redisRepository) Get(ctx context.Context, key string, v interface{}) error {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return fmt.Errorf("failed to get data from do: %v", err)
	}

	err = json.Unmarshal(reply, v)
	if err != nil {
		return fmt.Errorf("failed to get data: %v", err)
	}

	return nil
}

// MGet unmarshals values got from Redis to each elements of value `v`.
// The length of value `v` must be equal to the number elements of `keys`.
func (r *redisRepository) MGet(ctx context.Context, keys []interface{}, v interface{}) error {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	reply, err := redis.ByteSlices(conn.Do("MGET", keys...))
	if err != nil {
		return fmt.Errorf("failed to get data by MGET, err: %v", err)
	}

	// unmarshal to a slice of any literals
	p := reflect.ValueOf(v)
	s := reflect.Indirect(p)
	for i := 0; i < s.Len(); i++ {
		addr := s.Index(i).Addr().Interface()
		if err = json.Unmarshal(reply[i], addr); err != nil {
			return fmt.Errorf("failed to get data by MGET, err: %v", err)
		}
	}

	return nil
}

// Del unmarshals values got from Redis to each elements of value `v`.
// The length of value `v` must be equal to the number elements of `keys`.
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

// Exists returns whether `key` exists.
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

// SAdd adds a specified member to the set stored at key
func (r *redisRepository) SAdd(ctx context.Context, setGroup string, member string) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	num, err := redis.Int(conn.Do("SADD", setGroup, member))
	if num != 1 {
		return false, fmt.Errorf("failed to add member, setGroup: %s, member: %s", setGroup, member)
	}
	if err != nil {
		return false, fmt.Errorf("failed to add member, err: %v", err)
	}

	return true, nil
}

// SIsMember returns whether it is a member of the set stored at key.
func (r *redisRepository) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	conn, err := r.Pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("context expired. err: %v", err)
	}
	defer conn.Close()

	members, err := redis.Int(conn.Do("SISMEMBER", key, member))
	if err != nil {
		return false, fmt.Errorf("failed to search members in set: %v", err)
	}
	if members != 1 {
		return false, nil
	}

	return true, nil
}

// Keys returns all keys matching `pattern`.
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
