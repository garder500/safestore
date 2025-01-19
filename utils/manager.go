package utils

import (
	"context"
	"fmt"

	"github.com/garder500/safestore/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Manager struct {
	// GORM database connection
	DB               *gorm.DB
	pgx              *pgxpool.Pool
	Listener         *mapListener
	WebsocketManager *WebsocketManager
}

func NewManager() (*Manager, error) {
	// Set up GORM and pgx connections
	gormDB, pool, err := setupDB()
	if err != nil {
		return nil, fmt.Errorf("failed to set up database: %v", err)
	}

	// Create new MapListener
	if err := gormDB.Exec("CREATE EXTENSION IF NOT EXISTS ltree").Error; err != nil {
		return nil, err
	}

	// Auto migrate your models
	if err := gormDB.AutoMigrate(&database.SafeRow{}); err != nil {
		return nil, err
	}
	return &Manager{
		DB:               gormDB,
		pgx:              pool,
		Listener:         newMapListener(),
		WebsocketManager: NewWebsocketManager(),
	}, nil
}

func setupDB() (*gorm.DB, *pgxpool.Pool, error) {
	dsn := "host=localhost user=safeuser password=safepassword dbname=safestore port=5432 sslmode=disable TimeZone=Europe/Paris"

	// Set up GORM connection
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Set up pgx connection pool
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create connection pool: %v", err)
	}

	return gormDB, pool, nil
}

func (s *Manager) Notify(channel, payload string) error {
	return notify(s.pgx, channel, payload)
}

func notify(pool *pgxpool.Pool, channel, payload string) error {
	_, err := pool.Exec(context.Background(), "SELECT pg_notify($1, $2)", channel, payload)
	return err
}

func (s *Manager) ListenForNextPayload(channel string) (string, error) {
	return s.Listener.listenForNextPayload(channel)
}

func (s *Manager) Listen(channel string) error {
	conn, err := s.pgx.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	s.Listener.addChannel(channel)
	_, err = conn.Exec(context.Background(), "LISTEN "+channel)
	if err != nil {
		return fmt.Errorf("failed to start listening: %v", err)
	}

	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			return fmt.Errorf("error waiting for notification: %v", err)
		}
		fmt.Println("Received payload:", notification.Payload)
		s.Listener.Notify(channel, notification.Payload)
	}
}
