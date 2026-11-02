package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Zolet-hash/account-management-api/internal/config"
	"github.com/Zolet-hash/account-management-api/internal/infrastructure/database"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10e9)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseURL, nil)
	if err != nil {
		return err
	}

	defer db.Close()
	sql, err := os.ReadFile("scripts/seed.sql")
	if err != nil {
		return err
	}
	
	_, err = db.Exec(ctx, string(sql))
	return err
}
