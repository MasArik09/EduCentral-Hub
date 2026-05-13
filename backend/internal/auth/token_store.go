package auth

import (
	"sync"
	"time"
)

type RefreshTokenRecord struct {
	UserID    uint
	ExpiresAt time.Time
}

type InMemoryTokenStore struct {
	mu            sync.RWMutex
	refreshTokens map[string]RefreshTokenRecord
	accessLogout  map[string]time.Time
}

func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		refreshTokens: make(map[string]RefreshTokenRecord),
		accessLogout:  make(map[string]time.Time),
	}
}

func (s *InMemoryTokenStore) StoreRefreshToken(token string, record RefreshTokenRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[token] = record
}

func (s *InMemoryTokenStore) GetRefreshToken(token string, now time.Time) (RefreshTokenRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.refreshTokens[token]
	if !ok {
		return RefreshTokenRecord{}, false
	}
	if now.After(record.ExpiresAt) {
		delete(s.refreshTokens, token)
		return RefreshTokenRecord{}, false
	}

	return record, true
}

func (s *InMemoryTokenStore) DeleteRefreshToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshTokens, token)
}

func (s *InMemoryTokenStore) SetAccessTokenLogout(jti string, blockedAfter time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessLogout[jti] = blockedAfter
}

func (s *InMemoryTokenStore) CheckAccessTokenLogout(jti string, now time.Time) (deny bool, inGrace bool) {
	if jti == "" {
		return false, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	blockedAfter, ok := s.accessLogout[jti]
	if !ok {
		return false, false
	}

	if now.After(blockedAfter) {
		delete(s.accessLogout, jti)
		return true, false
	}

	return false, true
}
