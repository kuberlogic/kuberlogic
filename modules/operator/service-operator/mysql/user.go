package mysql

import (
	"database/sql"
	"fmt"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
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

const (
	readOnly   = 2
	fullAccess = 18
)

func execQueries(conn *sql.DB, queries ...string) error {
	for _, query := range queries {
		fmt.Println("-----------------")
		fmt.Println(query)
		_, err := conn.Exec(query) // multistatement queries do not allowed due to possible sql injections
		if err != nil {
			return err
		}
	}
	return nil
}

func changePermissions(conn *sql.DB, name string, permissions []interfaces.Permission) error {
	var queries []string

	queries = append(
		queries,
		// FIXME: possible sql injection, need to use args
		fmt.Sprintf("REVOKE ALL PRIVILEGES, GRANT OPTION FROM '%s'@'%%';", name),
	)

	for _, perm := range permissions {
		var type_ string
		switch perm.Privilege {
		case interfaces.ReadOnlyPrivilege:
			type_ = "SELECT, SHOW VIEW"
		case interfaces.FullPrivilege:
			type_ = "ALL PRIVILEGES"
		}
		queries = append(
			queries,
			// FIXME: possible sql injection, need to use args
			fmt.Sprintf("GRANT %s on %s.* to %s;", type_, perm.Database, name),
		)
	}

	queries = append(
		queries,
		"FLUSH PRIVILEGES;",
	)
	return execQueries(conn, queries...)
}

func editPassword(conn *sql.DB, name, password string) error {
	return execQueries(conn,
		fmt.Sprintf("ALTER USER '%s'@'%%' IDENTIFIED BY '%s';", name, password),
		"FLUSH PRIVILEGES;",
	)
}

func (usr *User) IsProtected(name string) bool {
	_, ok := protectedUsers[name]
	return ok
}

func (usr *User) Create(name, password string, permissions []interfaces.Permission) error {
	conn, err := sql.Open("mysql", usr.session.ConnectionString(usr.session.MasterIP, ""))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open doesn't open a connection. Validate DSN data:
	if err = conn.Ping(); err != nil {
		return err
	}

	if err = execQueries(
		conn,
		fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s';", name, password),
	); err != nil {
		return err
	}
	return changePermissions(conn, name, permissions)
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
	_, err = conn.Exec(fmt.Sprintf("DROP USER '%s'@'%%';", name))
	return err
}

func (usr *User) Edit(name, password string, permissions []interfaces.Permission) error {
	conn, err := sql.Open("mysql", usr.session.ConnectionString(usr.session.MasterIP, ""))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open doesn't open a connection. Validate DSN data:
	if err = conn.Ping(); err != nil {
		return err
	}

	if password != "" {
		err = editPassword(conn, name, password)
		if err != nil {
			return err
		}
	}
	return changePermissions(conn, name, permissions)
}

func (usr *User) List() (interfaces.Users, error) {
	var users = make(interfaces.Users)
	conn, err := sql.Open("mysql", usr.session.ConnectionString(usr.session.MasterIP, ""))
	if err != nil {
		return users, err
	}
	defer conn.Close()

	rows, err := conn.Query(`
SELECT u.user as name, sp.table_schema, COUNT(sp.privilege_type) as amount 
FROM mysql.user AS u 
LEFT JOIN information_schema.schema_privileges AS sp ON sp.grantee = concat('\'', u.user, '\'@\'', u.host, '\'') 
-- WHERE sp.table_schema IS NOT NULL 
GROUP BY u.user, sp.table_schema;
`)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		var db *string
		var amountOfPrivileges int
		err = rows.Scan(&username, &db, &amountOfPrivileges)
		if err != nil {
			return users, err
		}
		if !usr.IsProtected(username) {
			var privType interfaces.PrivilegeType
			switch amountOfPrivileges {
			case fullAccess:
				privType = interfaces.FullPrivilege
			case readOnly:
				privType = interfaces.ReadOnlyPrivilege
			default:
				privType = interfaces.UnknownPrivilege
			}

			// escape NULL values and using only grants which we could specified
			if db != nil && privType != interfaces.UnknownPrivilege {
				users[username] = append(users[username], interfaces.Permission{
					Database:  *db,
					Privilege: privType,
				})
			} else if db == nil {
				// user without any permissions
				users[username] = []interfaces.Permission{}
			}

		}
	}
	return users, nil
}
