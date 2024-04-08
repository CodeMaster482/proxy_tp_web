package postgresql

import (
	"context"
	"fmt"
	"proxy/pkg/config"

	"github.com/jackc/pgx/v4/pgxpool"
)

func InitPostgresDB(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Name, cfg.Database.Password, cfg.Database.Ssl)

	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
