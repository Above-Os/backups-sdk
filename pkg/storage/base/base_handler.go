package base

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

type Interface interface {
	SetOptions(opts *restic.ResticOptions)
	Backup(ctx context.Context) (backupSummary *restic.SummaryOutput, err error)
	Restore(ctx context.Context) (restoreSummary *restic.RestoreSummaryOutput, err error)
	Snapshots(ctx context.Context) (err error)
}
