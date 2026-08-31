package otp

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("token not found")
	ErrExpired  = errors.New("token expired")
)

// TokenMetadata tracks retry attempts and expiration.
type TokenMetadata struct {
	Code      string
	ExpiresAt time.Time
	Attempts  int
}

// Store defines the interface for persisting verification codes.
type Store interface {
	Set(ctx context.Context, key string, meta TokenMetadata) error
	Get(ctx context.Context, key string) (TokenMetadata, error)
	IncrementAttempts(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
}

// memoryStore is a thread-safe, in-memory implementation of Store.
type memoryStore struct {
	mu     sync.RWMutex
	tokens map[string]TokenMetadata
}

// NewMemoryStore creates a new in-memory Store.
func NewMemoryStore() Store {
	return &memoryStore{
		tokens: make(map[string]TokenMetadata),
	}
}

func (s *memoryStore) Set(ctx context.Context, key string, meta TokenMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[key] = meta
	return nil
}

func (s *memoryStore) Get(ctx context.Context, key string) (TokenMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	meta, ok := s.tokens[key]
	if !ok {
		return TokenMetadata{}, ErrNotFound
	}
	if time.Now().After(meta.ExpiresAt) {
		// Clean up dynamically
		go func() { _ = s.Delete(context.Background(), key) }()
		return TokenMetadata{}, ErrExpired
	}
	return meta, nil
}

func (s *memoryStore) IncrementAttempts(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, ok := s.tokens[key]
	if !ok {
		return ErrNotFound
	}
	meta.Attempts++
	s.tokens[key] = meta
	return nil
}

func (s *memoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, key)
	return nil
}
