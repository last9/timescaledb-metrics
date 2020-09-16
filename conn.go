package main

import "github.com/jackc/pgx"

func dbConn(dbURL string) (*pgx.Conn, error) {
	var cfg pgx.ConnConfig
	var err error

	if dbURL != "" {
		cfg, err = pgx.ParseConnectionString(dbURL)
	} else if err == nil {
		cfg, err = pgx.ParseEnvLibpq()
	}

	if err != nil {
		return nil, err
	}

	return pgx.Connect(cfg)
}
