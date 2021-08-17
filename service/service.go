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

type Service interface {
	Encode(ctx context.Context, fullUrl string, expiry *time.Time) (string, error)
	// Decode(shortCode string) (string, error)
	// GetUrlObjects(shortCode *string, fullUrl *string) ([]*model.UrlObject, error)
	// DeleteUrl(url string) (bool, error)
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
	return shortUrl, nil
}
