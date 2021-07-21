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

type dbUserPermission struct {
	username           string
	database           string
	amountOfPrivileges int
}

const (
	readOnly   = 2
	fullAccess = 18
)

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
		var permission dbUserPermission
		err = rows.Scan(&permission)
		if err != nil {
			return users, err
		}
		if !usr.IsProtected(permission.username) {
			var privType interfaces.PrivilegeType
			if permission.amountOfPrivileges == fullAccess {
				privType = interfaces.Full
			} else if permission.amountOfPrivileges == readOnly {
				privType = interfaces.ReadOnly
			}
			users[permission.username] = append(users[permission.username], interfaces.Permission{
				Database:  permission.database,
				Privilege: privType,
			})
		}
	}
	return users, nil
}
