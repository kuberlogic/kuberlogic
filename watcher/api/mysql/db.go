package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	session *Session
}

func (db *Database) Create(name string) error {
	conn, err := sql.Open("mysql", db.session.ConnectionString(db.session.MasterIP, ""))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open doesn't open a connection. Validate DSN data:
	if err = conn.Ping(); err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE DATABASE %s;", name)
	_, err = conn.Exec(query)
	return err
}

func (db *Database) Drop(name string) error {
	conn, err := sql.Open("mysql", db.session.ConnectionString(db.session.MasterIP, ""))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open doesn't open a connection. Validate DSN data:
	if err = conn.Ping(); err != nil {
		return err
	}

	query := fmt.Sprintf("DROP DATABASE %s;", name)

	_, err = conn.Exec(query)
	return err
}

func (db *Database) List() ([]string, error) {
	var names []string
	conn, err := sql.Open("mysql", db.session.ConnectionString(db.session.MasterIP, ""))
	if err != nil {
		return names, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
SELECT schema_name FROM information_schema.schemata;
`)
	if err != nil {
		return names, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return names, err
		}
		names = append(names, name)
	}
	return names, nil
}
