package filesystem

import (
	"path"

	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
)

type Filesystem struct {
	RepoName   string
	SnapshotId string
	Endpoint   string
	Password   string
	Path       string
}

// Backup implements storage.Location.
func (f *Filesystem) Backup() error {
	panic("unimplemented")
}

// Restore implements storage.Location.
func (f *Filesystem) Restore() error {
	panic("unimplemented")
}

// Snapshots implements storage.Location.
func (f *Filesystem) Snapshots() error {
	panic("unimplemented")
}

func (f *Filesystem) GetEnv(repository string) *restic.ResticEnv {
	var envs = &restic.ResticEnv{
		RESTIC_REPOSITORY: path.Join(f.Endpoint, f.RepoName),
		RESTIC_PASSWORD:   f.Password,
	}
	return envs
}

func (f *Filesystem) FormatRepository() (repository string, err error) {
	if err := f.setRepoDir(); err != nil {
		return "", err
	}
	return f.Endpoint, nil
}

func (f *Filesystem) GetRepoName() string {
	return f.RepoName
}

func (f *Filesystem) GetPath() string {
	return f.Path
}

func (f *Filesystem) GetSnapshotId() string {
	return f.SnapshotId
}

func (f *Filesystem) GetLimitUploadRate() string {
	return ""
}

func (f *Filesystem) GetLocation() common.Location {
	return common.LocationFileSystem
}

func (f *Filesystem) setRepoDir() error {
	var p = path.Join(f.Endpoint, f.RepoName)
	if !util.IsExist(p) {
		if err := util.CreateDir(p); err != nil {
			return err
		}
		return nil
	}
	return nil
}
