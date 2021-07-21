package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/kuberlogic/operator/modules/operator/service-operator/interfaces"
)

type User struct {
	session *Session
}

var protectedUsers = map[string]bool{
	"postgres":   true,
	"standby":    true,
	"kuberlogic": true,
}

func (usr *User) IsProtected(name string) bool {
	_, ok := protectedUsers[name]
	return ok
}

func (usr *User) Create(name, password string) error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, usr.session.ConnectionString(usr.session.MasterIP, "postgres"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf(`
CREATE USER %s WITH ENCRYPTED PASSWORD '%s';
ALTER USER %s WITH SUPERUSER;
`, name, password, name)
	_, err = conn.Exec(ctx, query)
	return err
}

func (usr *User) Delete(name string) error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, usr.session.ConnectionString(usr.session.MasterIP, "postgres"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf(`
DROP USER %s;
`, name)
	_, err = conn.Exec(ctx, query)
	return err
}

func (usr *User) Edit(name, password string) error {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, usr.session.ConnectionString(usr.session.MasterIP, "postgres"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	query := fmt.Sprintf(`
ALTER USER %s WITH PASSWORD '%s';
`, name, password)
	_, err = conn.Exec(ctx, query)
	return err
}

func (usr *User) List() (interfaces.Users, error) {
	var users = make(interfaces.Users)

	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, usr.session.ConnectionString(usr.session.MasterIP, "postgres"))
	if err != nil {
		return users, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
SELECT usename
FROM pg_catalog.pg_user;
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
		users[name] = append(users[name], interfaces.Permission{})
	}

	return users, nil
}
