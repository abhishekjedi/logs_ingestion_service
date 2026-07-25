// Package session issues and verifies the JWT that backs the dashboard login
// cookie. It is auth-source agnostic: dev-login and (later) Google both mint the
// same token once they've established a user id.
package session

import (
	"errors"
	"strconv"
	"time"

	"error-logging/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(cfg config.AuthConfig) *Manager {
	return &Manager{secret: []byte(cfg.JWTSecret), ttl: cfg.TokenTTL}
}

// Issue mints a signed token whose subject is the user id.
func (m *Manager) Issue(userID uint64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse validates a token and returns its user id.
func (m *Manager) Parse(token string) (uint64, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(claims.Subject, 10, 64)
}

// TTLSeconds is the cookie max-age.
func (m *Manager) TTLSeconds() int {
	return int(m.ttl.Seconds())
}
