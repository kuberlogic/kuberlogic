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
)

var protectedUsers = map[string]bool{
	"orchestrator":    true,
	"sys_operator":    true,
	"sys_replication": true,
	"sys_exporter":    true,
	"sys_heartbeat":   true,
	"mysql.sys":       true,
	"root":            true,
	"kuberlogic":      true,
}

type User struct {
	session *Session
}

func (usr *User) IsProtected(name string) bool {
	_, ok := protectedUsers[name]
	return ok
}

func (usr *User) Create(name, password string) error {
	conn, err := sql.Open("mysql", usr.session.ConnectionString(usr.session.MasterIP, ""))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open doesn't open a connection. Validate DSN data:
	if err = conn.Ping(); err != nil {
		return err
	}

	queries := []string{
		fmt.Sprintf("CREATE USER '%s'@'localhost' IDENTIFIED BY '%s';", name, password),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON *.* TO '%s'@'localhost';", name),
		"FLUSH PRIVILEGES;",
	}

	for _, query := range queries {
		_, err = conn.Exec(query) // multistatement queries do not allowed due to possible sql injections
		if err != nil {
			return err
		}
	}
	return nil
}

func (usr *User) Delete(name string) error {
	conn, err := sql.Open("mysql", usr.session.ConnectionString(usr.session.MasterIP, ""))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open doesn't open a connection. Validate DSN data:
	if err = conn.Ping(); err != nil {
		return err
	}
	_, err = conn.Exec(fmt.Sprintf("DROP USER '%s'@'localhost';", name))
	return err
}

func (usr *User) Edit(name, password string) error {
	conn, err := sql.Open("mysql", usr.session.ConnectionString(usr.session.MasterIP, ""))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open doesn't open a connection. Validate DSN data:
	if err = conn.Ping(); err != nil {
		return err
	}

	queries := []string{
		fmt.Sprintf("ALTER USER '%s'@'localhost' IDENTIFIED BY '%s';", name, password),
		"FLUSH PRIVILEGES;",
	}

	for _, query := range queries {
		_, err = conn.Exec(query) // multistatement queries do not allowed due to possible sql injections
		if err != nil {
			return err
		}
	}
	return nil
}

func (usr *User) List() ([]string, error) {
	var users []string
	conn, err := sql.Open("mysql", usr.session.ConnectionString(usr.session.MasterIP, ""))
	if err != nil {
		return users, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
SELECT user FROM mysql.user;
`)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return users, err
		}
		users = append(users, name)
	}
	return users, nil
}