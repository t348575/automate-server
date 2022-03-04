package config

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

func ProvidePostgres(config *Config) (*bun.DB, error) {
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(config.Dsn)))
	db := bun.NewDB(pgdb, pgdialect.New())
	if !config.IsProduction {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
		log.Info().Msg("Enabled bun debug")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
