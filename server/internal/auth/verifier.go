package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

const keyCacheTTL = time.Hour

var (
	ErrMissingToken = errors.New("missing auth token")
	ErrInvalidToken = errors.New("invalid token")
)

type Verifier interface {
	Verify(ctx context.Context, token string) (string, error)
}

type ClerkVerifier struct {
	parties     map[string]struct{}
	enforceAZP  bool
	mu          sync.Mutex
	keys        map[string]*clerk.JSONWebKey
	keysFetched time.Time
}

func NewClerkVerifier(secretKey string, authorizedParties []string) *ClerkVerifier {
	clerk.SetKey(secretKey)

	parties := make(map[string]struct{}, len(authorizedParties))
	for _, party := range authorizedParties {
		if party = strings.TrimSpace(party); party != "" {
			parties[party] = struct{}{}
		}
	}

	return &ClerkVerifier{
		parties:    parties,
		enforceAZP: len(parties) > 0,
		keys:       map[string]*clerk.JSONWebKey{},
	}
}

func (v *ClerkVerifier) Verify(ctx context.Context, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ErrMissingToken
	}

	keyID, err := tokenKeyID(token)
	if err != nil {
		return "", ErrInvalidToken
	}

	key, err := v.key(ctx, keyID)
	if err != nil {
		return "", err
	}

	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token:                  token,
		JWK:                    key,
		AuthorizedPartyHandler: v.authorizedParty,
	})
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", errors.New("invalid token: missing subject")
	}

	return claims.Subject, nil
}

func (v *ClerkVerifier) authorizedParty(azp string) bool {
	if !v.enforceAZP {
		return true
	}
	_, ok := v.parties[azp]
	return ok
}

func (v *ClerkVerifier) key(ctx context.Context, keyID string) (*clerk.JSONWebKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if cached, ok := v.keys[keyID]; ok && time.Since(v.keysFetched) < keyCacheTTL {
		return cached, nil
	}

	set, err := jwks.Get(ctx, &jwks.GetParams{})
	if err != nil {
		return nil, err
	}

	fetched := make(map[string]*clerk.JSONWebKey, len(set.Keys))
	for _, key := range set.Keys {
		if key != nil && key.KeyID != "" {
			fetched[key.KeyID] = key
		}
	}

	v.keys = fetched
	v.keysFetched = time.Now()

	if cached, ok := fetched[keyID]; ok {
		return cached, nil
	}

	return nil, errors.New("missing json web key")
}

func tokenKeyID(token string) (string, error) {
	segments := strings.Split(token, ".")
	if len(segments) < 2 {
		return "", ErrInvalidToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		return "", ErrInvalidToken
	}

	var header struct {
		KeyID string `json:"kid"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return "", ErrInvalidToken
	}
	if header.KeyID == "" {
		return "", ErrInvalidToken
	}

	return header.KeyID, nil
}
