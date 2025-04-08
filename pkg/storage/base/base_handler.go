package base

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

type Interface interface {
	SetOptions(opts *restic.ResticOptions)
	Backup(ctx context.Context, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, err error)
	Restore(ctx context.Context, progressCallback func(percentDone float64)) (restoreSummary *restic.RestoreSummaryOutput, err error)
	Snapshots(ctx context.Context) (err error)
	Stats(ctx context.Context) (*restic.StatsContainer, error)
}
