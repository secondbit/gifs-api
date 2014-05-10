package api

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"code.google.com/p/google-api-go-client/googleapi"
	"code.google.com/p/google-api-go-client/oauth2/v2"
)

var (
	InvalidBearerToken = errors.New("Invalid bearer token")
)

type Authorizer interface {
	Authorize(token string, c Context) (string, error)
}

type GoogleOAuth2Authorizer struct {
	ClientID string
	cache    map[string]struct {
		user_id string
		expires time.Time
	}
	sync.Mutex
}

func NewGoogleOAuth2Authorizer(id string) *GoogleOAuth2Authorizer {
	return &GoogleOAuth2Authorizer{
		ClientID: id,
		cache: make(map[string]struct {
			user_id string
			expires time.Time
		}),
	}
}

func (g *GoogleOAuth2Authorizer) Authorize(token string, c Context) (string, error) {
	g.Lock()
	defer g.Unlock()
	if info, ok := g.cache[token]; ok && time.Now().Before(info.expires) {
		return info.user_id, nil
	} else if ok {
		delete(g.cache, token)
	}
	service, err := oauth2.New(http.DefaultClient)
	if err != nil {
		return "", err
	}
	info, err := service.Tokeninfo().Access_token(token).Do()
	if err != nil {
		if gErr, ok := err.(*googleapi.Error); ok {
			if gErr.Code >= 400 && gErr.Code < 500 {
				return "", InvalidBearerToken
			}
		}
		return "", err
	}
	if info.Audience != g.ClientID {
		return "", InvalidBearerToken
	}
	g.cache[token] = struct {
		user_id string
		expires time.Time
	}{
		user_id: info.User_id,
		expires: time.Now().Add(15 * time.Minute),
	}
	return info.User_id, nil
}
