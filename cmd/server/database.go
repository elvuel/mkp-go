package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	servercfg "github.com/elvuel/mkp-go/cmd/server/config"
	"github.com/elvuel/mkp-go/cmd/server/models"
	gormMySQL "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initDatabase(cfg servercfg.DatabaseConfig) (*gorm.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	dsn := strings.TrimSpace(cfg.DSN)

	var db *gorm.DB
	var err error

	switch driver {
	case "", "sqlited", "sqlite", "sqlite3":
		if dsn == "" {
			dsn = "mkp.db"
		}
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})

	case "mysql":
		if dsn == "" {
			return nil, fmt.Errorf("mysql requires a non-empty DSN")
		}
		db, err = gorm.Open(gormMySQL.Open(dsn), &gorm.Config{})

	case "postgres", "postgresql":
		if dsn == "" {
			return nil, fmt.Errorf("postgres requires a non-empty DSN")
		}
		db, err = gorm.Open(gormPostgres.Open(dsn), &gorm.Config{})

	default:
		return nil, fmt.Errorf("unsupported database driver: %s (supported: sqlited/sqlite/sqlite3, mysql, postgres)", cfg.Driver)
	}
	if err != nil {
		return nil, err
	}

	if err := applyPoolConfig(db, cfg); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.MacroRecord{}); err != nil {
		return nil, err
	}

	if cfg.WithOTEL {
		log.Printf("database.withOTEL=true (OTEL instrumentation is not wired in this server)")
	}
	log.Printf("database initialized: driver=%s dsn=%s", driver, dsn)
	return db, nil
}

func applyPoolConfig(db *gorm.DB, cfg servercfg.DatabaseConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxIdleTimeInSeconds != nil && *cfg.ConnMaxIdleTimeInSeconds > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(*cfg.ConnMaxIdleTimeInSeconds) * time.Second)
	}
	if cfg.ConnMaxLifetimeInSeconds != nil && *cfg.ConnMaxLifetimeInSeconds > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(*cfg.ConnMaxLifetimeInSeconds) * time.Second)
	}

	return nil
}
