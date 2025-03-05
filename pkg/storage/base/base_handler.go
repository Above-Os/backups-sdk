package base

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

type Interface interface {
	SetOptions(opts *restic.ResticOptions)
	Backup() (err error)
	Restore() (err error)
	Snapshots() (err error)
}
