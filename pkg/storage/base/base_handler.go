package base

import (
	"context"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

type Interface interface {
	SetOptions(opts *restic.ResticOptions)
	Backup(ctx context.Context, dryRun bool, progressCallback func(percentDone float64)) (backupSummary *restic.SummaryOutput, err error)
	Restore(ctx context.Context, progressCallback func(percentDone float64)) (map[string]*restic.RestoreSummaryOutput, string, uint64, error)
	Snapshots(ctx context.Context) (*restic.SnapshotList, error)
	Stats(ctx context.Context) (*restic.StatsContainer, error)
}
