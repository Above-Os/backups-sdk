package storage

type Location interface {
	Backup() error
	Restore() error
}
