package otp

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"
)

var (
	ErrMaxRetriesExceeded = errors.New("maximum verification attempts exceeded")
	ErrInvalidCode        = errors.New("verification code is invalid")
)

// CodeGenerator defines a function signature for custom verification code generators.
type CodeGenerator func(length int) (string, error)

// Config represents settings for the OTP verification engine.
type Config struct {
	// CodeLength specifies the length of the OTP (default is 6).
	CodeLength int
	// TTL specifies how long the code is valid (default is 5 minutes).
	TTL time.Duration
	// MaxRetries specifies how many attempts are allowed before code invalidation (default is 3).
	MaxRetries int
	// Generator specifies a custom code generator. If nil, a cryptographically secure numeric generator is used.
	Generator CodeGenerator
}

// Keeper coordinates OTP request and verification flows.
type Keeper struct {
	store     Store
	deliverer Deliverer
	cfg       Config
}

// New creates a new instance of Keeper.
func New(store Store, deliverer Deliverer, cfg Config) *Keeper {
	if cfg.CodeLength == 0 {
		cfg.CodeLength = 6
	}
	if cfg.TTL == 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.Generator == nil {
		cfg.Generator = SecureNumericGenerator
	}

	return &Keeper{
		store:     store,
		deliverer: deliverer,
		cfg:       cfg,
	}
}

// Knock generates a verification code, saves it to the store, and delivers it.
func (k *Keeper) Knock(ctx context.Context, recipient string) error {
	code, err := k.cfg.Generator(k.cfg.CodeLength)
	if err != nil {
		return err
	}

	meta := TokenMetadata{
		Code:      code,
		ExpiresAt: time.Now().Add(k.cfg.TTL),
		Attempts:  0,
	}

	if err := k.store.Set(ctx, recipient, meta); err != nil {
		return err
	}

	return k.deliverer.Deliver(ctx, recipient, code)
}

// Verify checks the provided code against the store, handles retry counts, and cleans up on success.
func (k *Keeper) Verify(ctx context.Context, recipient string, code string) (bool, error) {
	meta, err := k.store.Get(ctx, recipient)
	if err != nil {
		return false, err
	}

	// Increment attempts first
	if err := k.store.IncrementAttempts(ctx, recipient); err != nil {
		return false, err
	}
	meta.Attempts++

	// Verify maximum attempts
	if meta.Attempts >= k.cfg.MaxRetries {
		_ = k.store.Delete(ctx, recipient)
		return false, ErrMaxRetriesExceeded
	}

	// Compare codes
	if meta.Code != code {
		return false, ErrInvalidCode
	}

	// Success! Delete token and allow entry
	_ = k.store.Delete(ctx, recipient)
	return true, nil
}

// SecureNumericGenerator generates a cryptographically secure string of digits.
func SecureNumericGenerator(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}
	return string(result), nil
}
