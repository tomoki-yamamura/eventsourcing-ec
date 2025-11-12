package client

import (
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/config"
)

const (
	maxOpenConns    = 20
	maxIdleConns    = 20
	maxConnLifeTime = 5 * time.Minute
	maxConnIdleTime = 2 * time.Minute
)

type Client struct {
	DB *sqlx.DB
}

func NewClient(cfg config.DatabaseConfig) (*Client, error) {
	c := mysql.Config{
		User:                 cfg.User,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		DBName:               cfg.Name,
		ParseTime:            true,
		Collation:            "utf8mb4_unicode_ci",
		AllowNativePasswords: true,
	}

	c.Params = map[string]string{
		"charset": "utf8mb4",
	}

	dsn := c.FormatDSN()

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(maxConnLifeTime)
	db.SetConnMaxIdleTime(maxConnIdleTime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Client{DB: db}, nil
}

func (c *Client) GetDB() *sqlx.DB {
	return c.DB
}

func (c *Client) Close() error {
	return c.DB.Close()
}
