package base

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

type Interface interface {
	SetOptions(opts *restic.ResticOptions)
	Backup() (backupSummary *restic.SummaryOutput, err error)
	Restore() (restoreSummary *restic.RestoreSummaryOutput, err error)
	Snapshots() (err error)
}
