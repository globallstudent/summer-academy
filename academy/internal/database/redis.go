package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/globallstudent/academy/internal/config"
	"github.com/redis/go-redis/v9"
)

// Redis represents a Redis connection
type Redis struct {
	Client   *redis.Client
	fallback map[string]otpEntry
	mu       sync.Mutex
}

type otpEntry struct {
	Code   string
	Expiry time.Time
}

// ConnectRedis creates a new Redis connection
func ConnectRedis(cfg config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		// Return Redis struct with fallback storage
		return &Redis{Client: nil, fallback: make(map[string]otpEntry)},
			fmt.Errorf("unable to connect to Redis: %w", err)
	}

	return &Redis{Client: client, fallback: make(map[string]otpEntry)}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}

// StoreOTP stores a one-time password with expiration
func (r *Redis) StoreOTP(phoneNumber string, otp string, expiry time.Duration) error {
	if r.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		key := fmt.Sprintf("otp:%s", phoneNumber)
		if err := r.Client.Set(ctx, key, otp, expiry).Err(); err == nil {
			return nil
		}
	}
	// fallback storage
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback[phoneNumber] = otpEntry{Code: otp, Expiry: time.Now().Add(expiry)}
	return nil
}

// VerifyOTP checks if an OTP is valid and deletes it if it is
func (r *Redis) VerifyOTP(phoneNumber string, otp string) (bool, error) {
	if r.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		key := fmt.Sprintf("otp:%s", phoneNumber)
		storedOTP, err := r.Client.Get(ctx, key).Result()
		if err == nil {
			if storedOTP == otp {
				r.Client.Del(ctx, key)
				return true, nil
			}
			return false, nil
		}
	}

	// Fallback check
	r.mu.Lock()
	entry, ok := r.fallback[phoneNumber]
	if ok && time.Now().Before(entry.Expiry) && entry.Code == otp {
		delete(r.fallback, phoneNumber)
		r.mu.Unlock()
		return true, nil
	}
	if ok && time.Now().After(entry.Expiry) {
		delete(r.fallback, phoneNumber)
	}
	r.mu.Unlock()

	return false, nil
}
