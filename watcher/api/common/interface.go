package common

type Database interface {
	List() ([]string, error)
	Create(name string) error
	Drop(name string) error
}

type Session interface {
	String() string
	SetupDDL() error
	RunQueries(delay Delay, duration Duration)

	GetDatabase() Database
}
