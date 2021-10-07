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

package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
)

var protectedDatabases = map[string]bool{
	"postgres":   true,
	"kuberlogic": true,
}

type Database struct {
	session *Session
}

func (db *Database) IsProtected(name string) bool {
	_, ok := protectedDatabases[name]
	return ok
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
