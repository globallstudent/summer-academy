package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/academy/internal/config"
)

// Redis represents a Redis connection
type Redis struct {
	Client *redis.Client
}

// ConnectRedis creates a new Redis connection
func ConnectRedis(cfg config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	// Test the connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	return &Redis{Client: client}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}

// StoreOTP stores a one-time password with expiration
func (r *Redis) StoreOTP(phoneNumber string, otp string, expiry time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", phoneNumber)
	return r.Client.Set(ctx, key, otp, expiry).Err()
}

// VerifyOTP checks if an OTP is valid and deletes it if it is
func (r *Redis) VerifyOTP(phoneNumber string, otp string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", phoneNumber)

	// Get the stored OTP
	storedOTP, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // OTP not found
	} else if err != nil {
		return false, err
	}

	// Check if the OTP matches
	if storedOTP == otp {
		// Delete the OTP to prevent reuse
		if err := r.Client.Del(ctx, key).Err(); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}
