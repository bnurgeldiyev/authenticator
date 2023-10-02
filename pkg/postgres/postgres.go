package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"authenticator/config"
)

// Postgres -.
type Postgres struct {
	Builder sq.StatementBuilderType
	Pool    *pgxpool.Pool

	txMap   map[int]*connTx
	txMutex sync.RWMutex
	idTx    int
}

// Config struct holds all the configurations required the postgres package
type Config struct {
	Host   string
	Port   string
	Driver string

	StoreName  string
	Username   string
	Password   string
	SSLMode    string
	SearchPath string

	ConnPoolSize uint
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	DialTimeout  time.Duration
}

// ConnURL returns the connection Url
func (cfg *Config) ConnURL() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
		cfg.Driver,
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.StoreName,
		cfg.SSLMode,
		cfg.SearchPath,
	)
}

// NewService returns a new instance of PGX pool
func NewService(cfg *config.Config) (db *Postgres, err error) {
	ctx := context.Background()
	var conn *pgx.Conn
	conn, err = pgx.Connect(ctx, cfg.Database.URL)
	if err != nil {
		return nil, err
	}
	conn.Close(ctx)

	var poolCfg *pgxpool.Config
	poolCfg, err = pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = int32(cfg.Database.PoolMax)

	var pool *pgxpool.Pool
	pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		Pool:    pool,
		Builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		txMap:   make(map[int]*connTx, 100),
	}, nil
}
