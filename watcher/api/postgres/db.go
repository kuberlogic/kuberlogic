package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
)

type Database struct {
	session *Session
}

func (db *Database) Create(name string) error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, db.session.ConnectionString(db.session.MasterIP, "postgres"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf("CREATE DATABASE %s;", name)
	_, err = conn.Exec(ctx, query)
	return err
}

func (db *Database) Drop(name string) error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, db.session.ConnectionString(db.session.MasterIP, "postgres"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf("DROP DATABASE %s;", name)
	_, err = conn.Exec(ctx, query)
	return err
}

func (db *Database) List() ([]string, error) {
	var items []string

	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, db.session.ConnectionString(db.session.MasterIP, "postgres"))
	if err != nil {
		return items, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
SELECT datname FROM pg_database
WHERE datistemplate = false;`)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return items, err
		}
		items = append(items, name)
	}

	return items, nil
}
