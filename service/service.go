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

const shortenUrlPrefix = "shortenUrl:"
const fullUrlPrefix = "fullUrl:"
const deletedShortUrlKey = "deletedShortUrlKey"

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
	}

	shortCodeKey := shortenUrlPrefix + shortCode
	fullUrlKey := fullUrlPrefix + fullUrl

	_, err := s.repository.Set(ctx, shortCodeKey, object, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to set object, err: %v", err)
	}
	_, err = s.repository.Set(ctx, fullUrlKey, shortCodeKey, expiry)
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

	shortCodeKey := getShortCodeKey(shortCode)

	var object model.UrlObject
	err = s.repository.Get(ctx, shortenUrlPrefix+shortCode, &object)
	if err != nil {
		return "", fmt.Errorf("failed to get url, err: %v", err)
	}

	object.Hits += 1

	_, err = s.repository.Set(ctx, shortCodeKey, object, &object.Expiry)
	if err != nil {
		return "", fmt.Errorf("failed to set object, err: %v", err)
	}
	return object.FullURL, nil
}

func (s *service) GetUrlObjects(ctx context.Context, shortCode *string, fullUrl *string) ([]*model.UrlObject, error) {
	var shortCodeKeys []string
	var fullUrlKeys []string
	var err error

	// filter short code
	shortCodePattern := shortenUrlPrefix + "*"
	if shortCode != nil {
		shortCodePattern = shortenUrlPrefix + "*" + *shortCode + "*"
	}

	// search all possible shortCodeKeys
	shortCodeKeys, err = s.repository.Keys(ctx, shortCodePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get members, err: %v", err)
	}

	// filter full url
	if fullUrl != nil {
		fullUrlKeys, err = s.repository.Keys(ctx, fullUrlPrefix+"*"+*fullUrl+"*")
		if err != nil {
			return nil, fmt.Errorf("failed to get members, err: %v", err)
		}
		if len(fullUrlKeys) != 0 {
			var keys []interface{}
			for _, key := range fullUrlKeys {
				keys = append(keys, key)
			}
			var filteredShortCodeKeysByFullUrl []string
			for _ = range fullUrlKeys {
				filteredShortCodeKeysByFullUrl = append(filteredShortCodeKeysByFullUrl, "")
			}

			// get all shortCodes from all fullUrlKeys
			err := s.repository.MGet(ctx, keys, &filteredShortCodeKeysByFullUrl)
			if err != nil {
				return nil, fmt.Errorf("failed to get short codes from full url keys, err: %v", err)
			}

			fmt.Printf("after filteredShortCodeKeysByFullUrl: %v\n", filteredShortCodeKeysByFullUrl)

			// intersect with shortCodeKeys
			shortCodeKeys = intersect(shortCodeKeys, filteredShortCodeKeysByFullUrl)
		}

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
	isDeleted, err := s.repository.Del(ctx, shortenUrlPrefix+shortCode)
	if err != nil || !isDeleted {
		return false, fmt.Errorf("failed to delete url, err: %v", err)
	}
	_, err = s.repository.SAdd(ctx, deletedShortUrlKey, shortCode)
	if err != nil {
		return false, fmt.Errorf("failed to set object, err: %v", err)
	}
	return isDeleted, err
}

func getShortCodeKey(shortCode string) string {
	return shortenUrlPrefix + shortCode
}
func getFullUrlKey(fullUrl string) string {
	return fullUrlPrefix + fullUrl
}
func intersect(a []string, b []string) []string {
	var result []string
	set := make(map[string]bool)

	for _, v := range a {
		set[v] = true
	}
	for _, v := range b {
		if set[v] {
			result = append(result, v)
		}
	}
	return result
}
