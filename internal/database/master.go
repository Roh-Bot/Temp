package database

import (
	"context"
	"fmt"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Database struct {
	*pgxpool.Pool
}

func NewMasterConnection(config config.Database) (db *Database, err error) {
	connString := fmt.Sprintf(
		`host=%s port=%s user=%s password=%s database=%s sslmode=%s`,
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return
	}
	cfg.MaxConnIdleTime = config.MaxConnectionIdleTime
	cfg.MaxConnLifetime = config.MaxConnectionLifetime

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return
	}

	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Database{pool}, nil
}

func (d *Database) Flush() {
	d.Close()
}
