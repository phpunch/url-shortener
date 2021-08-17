package service

import (
	"context"
	"fmt"
	"github.com/catinello/base62"
	"math/rand"
	"strconv"
	"time"
	"url-shortener/model"
	"url-shortener/repository"
)

const memberKey = "shortenUrlMembers"

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
	exist := true
	for exist {
		id := rand.Uint64()
		exist, err = s.repository.Exists(ctx, strconv.FormatUint(id, 10))
		if err != nil || exist {
			continue
		}
	}
	return base62.Encode(id)
}

func (s *service) Encode(ctx context.Context, fullUrl string, expiry *time.Time) (string, error) {
	object := &model.UrlObject{
		FullURL: fullUrl,
		Hits:    0,
	}

	shortUrl := s.generateShortUrl(ctx)
	object.ShortCode = shortUrl

	if expiry != nil {
		object.Expiry = *expiry
	}

	_, err := s.repository.Set(ctx, object)
	if err != nil {
		return "", fmt.Errorf("failed to set object, err: %v", err)
	}
	_, err = s.repository.SAdd(ctx, memberKey, shortUrl)
	if err != nil {
		return "", fmt.Errorf("failed to add member, err: %v", err)
	}
	return shortUrl, nil
}

func (s *service) Decode(ctx context.Context, shortCode string) (string, error) {
	object, err := s.repository.Get(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get url, err: %v", err)
	}

	object.Hits += 1

	_, err = s.repository.Set(ctx, object)
	if err != nil {
		return "", fmt.Errorf("failed to set object, err: %v", err)
	}
	return object.FullURL, nil
}

func (s *service) GetUrlObjects(ctx context.Context, shortCode *string, fullUrl *string) ([]*model.UrlObject, error) {
	shortCodes, err := s.repository.SMembers(ctx, memberKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get members, err: %v", err)
	}

	// TODO: filter

	var urlObjects []*model.UrlObject
	for _, shortCode := range shortCodes {
		urlObject, err := s.repository.Get(ctx, shortCode)
		if err != nil {
			return nil, fmt.Errorf("failed to get url, err: %v", err)
		}
		urlObjects = append(urlObjects, urlObject)
	}

	return urlObjects, nil
}
