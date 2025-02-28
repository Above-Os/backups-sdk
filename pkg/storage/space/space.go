package space

import "bytetrade.io/web3os/backups-sdk/pkg/restic"

type Space struct {
	RepoName       string
	SnapshotId     string
	RepoRegion     string
	OlaresId       string
	Password       string
	Path           string
	CloudApiMirror string
	BaseDir        string
}

type StorageResponse struct {
	Summary          *restic.SummaryOutput
	RestoreSummary   *restic.RestoreSummaryOutput
	SnapshotsSummary []*restic.Snapshot
	Error            error
}
