package storage

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
)

type StorageClient struct {
	RepoName        string
	SnapshotId      string
	OlaresId        string
	StorageLocation string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string

	UploadPath        string
	TargetPath        string
	CloudApiMirror    string
	TokenDuration     string
	LimitUploadRate   string
	LimitDownloadRate string

	BaseDir string
	Version string
}

type StorageResponse struct {
	Summary          *restic.SummaryOutput
	RestoreSummary   *restic.RestoreSummaryOutput
	SnapshotsSummary []*restic.Snapshot
	Error            error
}

type BackupsOperate string

var (
	OperateBackup    BackupsOperate = "backup"
	OperateRestore   BackupsOperate = "restore"
	OperateSnapshots BackupsOperate = "snapshots"
)

func (o BackupsOperate) IsBackup() bool {
	return o == OperateBackup
}
