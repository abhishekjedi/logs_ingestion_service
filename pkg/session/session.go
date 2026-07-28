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

func (m *Manager) Issue(userID uint64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

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

func (m *Manager) TTLSeconds() int {
	return int(m.ttl.Seconds())
}
