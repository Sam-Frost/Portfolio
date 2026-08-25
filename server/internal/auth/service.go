// Package auth gates the domain area behind a single shared password (set
// via DOMAIN_PASSWORD) and issues short-lived JWTs (signed with JWT_SECRET)
// on successful login. There are no per-user accounts — this mirrors the
// existing "one password gets you into /dashboard" model already implied by
// DomainExpansionPage, just enforced server-side now.
package auth

import (
	"crypto/subtle"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

const tokenTTL = 24 * time.Hour

type Service struct {
	password string
	secret   []byte
}

func NewService(password, secret string) *Service {
	return &Service{password: password, secret: []byte(secret)}
}

// Login checks password against the configured DOMAIN_PASSWORD (constant
// time, since this is the entire access control for the domain area) and
// returns a signed JWT on success.
func (s *Service) Login(password string) (string, error) {
	if subtle.ConstantTimeCompare([]byte(password), []byte(s.password)) != 1 {
		return "", apperr.Unauthorized("incorrect password")
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   "domain",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", apperr.Internal("failed to sign token")
	}
	return signed, nil
}

// Verify returns nil if tokenString is a currently-valid token issued by
// Login, and an *apperr.Error (Unauthorized) otherwise.
func (s *Service) Verify(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperr.Unauthorized("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return apperr.Unauthorized("invalid or expired token")
	}
	return nil
}
