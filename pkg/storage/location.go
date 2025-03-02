package storage

import "bytetrade.io/web3os/backups-sdk/pkg/restic"

type Location interface {
	Backup() error
	Restore() error
	GetEnv(repository string) *restic.ResticEnv
	FormatRepository() (repository string, err error)
}
