package service

import (
	"context"
	"fmt"
	"github.com/catinello/base62"
	"math/rand"
	"net/http"
	"time"
	"url-shortener/customError"
	"url-shortener/model"
	"url-shortener/repository"
)

const deletedShortUrlKey = "deletedShortUrlKey"
const keyPattern = "url:%s#%s"

var maxTime, _ = time.Parse(time.RFC3339, "9999-12-31T23:59:59+07:00")

type Service interface {
	Encode(ctx context.Context, fullUrl string, expiry *time.Time) (string, error)
	Decode(ctx context.Context, shortCode string) (string, error)
	GetUrlObjects(ctx context.Context, shortCode *string, fullUrl *string) ([]*model.UrlObject, error)
	DeleteUrl(ctx context.Context, url string) (bool, error)
}

type service struct {
	repository repository.Repository
}

func New(repo repository.Repository) Service {
	service := &service{
		repository: repo,
	}
	return service
}

func (s *service) generateShortUrl(ctx context.Context) string {
	var id int
	var err error
	var key string
	exist := true
	for exist {
		id = rand.Int()
		key = base62.Encode(id)
		exist, err = s.repository.Exists(ctx, key)
		if err != nil || exist {
			continue
		}
		exist, err = s.repository.SIsMember(ctx, deletedShortUrlKey, key)
		if err != nil || exist {
			continue
		}
	}
	return key
}

func (s *service) Encode(ctx context.Context, fullUrl string, expiry *time.Time) (string, error) {
	object := &model.UrlObject{
		FullURL: fullUrl,
		Hits:    0,
	}

	shortCode := s.generateShortUrl(ctx)
	object.ShortCode = shortCode

	if expiry != nil {
		object.Expiry = *expiry
	} else {
		// set maxTime to expiry if a user don't specify this field
		object.Expiry = maxTime
	}

	shortCodeKey := fmt.Sprintf(keyPattern, shortCode, fullUrl)
	_, err := s.repository.Set(ctx, shortCodeKey, object, &object.Expiry)
	if err != nil {
		return "", fmt.Errorf("failed to set object, err: %v", err)
	}

	return shortCode, nil
}

func (s *service) Decode(ctx context.Context, shortCode string) (string, error) {
	deleted, err := s.repository.SIsMember(ctx, deletedShortUrlKey, shortCode)
	if err != nil {
		return "", err
	}
	if deleted {
		return "", &customError.InternalError{
			Code:           0,
			Message:        "this short code is already deleted",
			HTTPStatusCode: http.StatusGone,
		}
	}

	shortCodeKey := fmt.Sprintf(keyPattern, shortCode, "*")

	keys, err := s.repository.Keys(ctx, shortCodeKey)
	if err != nil {
		return "", fmt.Errorf("failed to get url, err: %v", err)
	}

	if len(keys) != 1 {
		return "", fmt.Errorf("failed to get url, err: not single key, keys: %v", keys)
	}

	var object model.UrlObject
	err = s.repository.Get(ctx, keys[0], &object)
	if err != nil {
		return "", fmt.Errorf("failed to get url, err: %v", err)
	}

	object.Hits += 1

	_, err = s.repository.Set(ctx, keys[0], object, &object.Expiry)
	if err != nil {
		return "", fmt.Errorf("failed to set object, err: %v", err)
	}
	return object.FullURL, nil
}

func (s *service) GetUrlObjects(ctx context.Context, shortCode *string, fullUrl *string) ([]*model.UrlObject, error) {
	var shortCodeKeys []string
	// var fullUrlKeys []string
	var err error

	// filter short code
	shortCodePattern := fmt.Sprintf(keyPattern, "*", "*")
	if shortCode != nil && fullUrl != nil {
		shortCodePattern = fmt.Sprintf(keyPattern, "*"+*shortCode+"*", "*"+*fullUrl+"*")
	} else if shortCode != nil {
		shortCodePattern = fmt.Sprintf(keyPattern, "*"+*shortCode+"*", "*")
	} else if fullUrl != nil {
		shortCodePattern = fmt.Sprintf(keyPattern, "*", "*"+*fullUrl+"*")
	}

	// search all possible shortCodeKeys
	shortCodeKeys, err = s.repository.Keys(ctx, shortCodePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get members, err: %v", err)
	}

	var urlObjects []*model.UrlObject
	for _, shortCodeKey := range shortCodeKeys {
		var urlObject model.UrlObject
		err := s.repository.Get(ctx, shortCodeKey, &urlObject)
		if err != nil {
			return nil, fmt.Errorf("failed to get url, err: %v", err)
		}
		urlObjects = append(urlObjects, &urlObject)
	}

	return urlObjects, nil
}

func (s *service) DeleteUrl(ctx context.Context, shortCode string) (bool, error) {
	shortCodeKey := fmt.Sprintf(keyPattern, shortCode, "*")

	keys, err := s.repository.Keys(ctx, shortCodeKey)
	if err != nil {
		return false, fmt.Errorf("failed to get key, err: %v", err)
	}
	if len(keys) != 1 {
		return false, fmt.Errorf("failed to get key, err: not single key, keys: %v", keys)
	}

	isDeleted, err := s.repository.Del(ctx, keys[0])
	if err != nil || !isDeleted {
		return false, fmt.Errorf("failed to delete url, err: %v", err)
	}
	_, err = s.repository.SAdd(ctx, deletedShortUrlKey, shortCode)
	if err != nil {
		return false, fmt.Errorf("failed to set object, err: %v", err)
	}
	return isDeleted, err
}
