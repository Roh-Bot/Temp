package database

import (
	"context"
	"fmt"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache struct {
	*redis.Client
}

func NewCache(cfg *config.AtomicConfig) (*Cache, error) {
	c := cfg.Get().Cache
	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprint(c.Host, ":", c.Port),
		OnConnect:       nil,
		Protocol:        c.Protocol,
		Username:        c.User,
		Password:        c.Password,
		DB:              c.Db,
		MaxRetries:      c.MaxRetries,
		MinRetryBackoff: c.MinRetryBackoff,
		MaxRetryBackoff: c.MaxRetryBackoff,
		DialTimeout:     c.DialTimeout,
		ReadTimeout:     c.ReadTimeout,
		WriteTimeout:    c.WriteTimeout,
		PoolFIFO:        c.PoolFifo,
		PoolSize:        c.PoolSize,
		PoolTimeout:     c.PoolTimeout,
		MinIdleConns:    c.MinIdleConns,
		ConnMaxIdleTime: c.ConnMaxIdleTime,
		ConnMaxLifetime: c.ConnMaxLifetime,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Cache{client}, nil
}
