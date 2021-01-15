package common

type Database interface {
	List() ([]string, error)
	Create(name string) error
	Drop(name string) error
}

type User interface {
	List() ([]string, error)
	Create(name, password string) error
	Delete(name string) error
	Edit(name, password string) error
}

type Session interface {
	String() string
	SetupDDL() error
	RunQueries(delay Delay, duration Duration)
	GetDatabase() Database
	GetUser() User
}
