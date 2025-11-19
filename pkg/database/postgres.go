package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshbarros/golang-airport-services/pkg/config"
	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"go.uber.org/zap"
)

// PostgresDB wraps pgxpool for database operations
type PostgresDB struct {
	Pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(cfg *config.Config, log *logger.Logger) (*PostgresDB, error) {
	// Build connection string
	dsn := cfg.GetDatabaseDSN()

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Set pool configuration based on config
	poolConfig.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.Database.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.Database.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.ConnMaxIdleTime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Info("Connected to PostgreSQL",
		zap.String("host", cfg.Database.Host),
		zap.String("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name),
		zap.Int("max_conns", cfg.Database.MaxOpenConns),
	)

	return &PostgresDB{
		Pool:   pool,
		logger: log,
	}, nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("Closed PostgreSQL connection pool")
	}
}

// Ping checks if the database connection is alive
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Health returns database health status
func (db *PostgresDB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return db.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// WithTransaction executes a function within a transaction
func (db *PostgresDB) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			db.logger.Error("Failed to rollback transaction",
				zap.Error(err),
				zap.Error(rbErr),
			)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
