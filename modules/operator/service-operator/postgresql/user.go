package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
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

func (usr *User) List() ([]string, error) {
	var items []string

	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, usr.session.ConnectionString(usr.session.MasterIP, "postgres"))
	if err != nil {
		return items, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
SELECT usename
FROM pg_catalog.pg_user;
`)
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
