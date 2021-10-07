/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var protectedDatabases = map[string]bool{
	"information_schema": true,
	"mysql":              true,
	"performance_schema": true,
	"sys":                true,
	"sys_operator":       true,
}

type Database struct {
	session *Session
}

func (db *Database) IsProtected(name string) bool {
	_, ok := protectedDatabases[name]
	return ok
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
