package storage

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
)

type SnapshotsProvider interface {
	Snapshots() error
}

type Snapshots struct {
	option SnapshotsOption
	client *StorageClient
}

type SnapshotsOption struct {
	RepoName        string `json:"repo_name"`
	OlaresId        string `json:"olares_id"`
	StorageLocation string `json:"storage_location"`
	Endpoint        string `json:"endpoint"`
	AccessKey       string `json:"access_key"`
	SecretAccessKey string `json:"secret_access_key"`
	CloudApiMirror  string `json:"cloud_api_mirror"`
	BaseDir         string `json:"base_dir"`
	Version         string `json:"version"`
}

func (u *Snapshots) Snapshots(opt SnapshotsOption) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	u.option = opt

	var storageClient = &StorageClient{
		RepoName:        u.option.RepoName,
		OlaresId:        u.option.OlaresId,
		StorageLocation: u.option.StorageLocation,
		Endpoint:        u.option.Endpoint,
		AccessKey:       u.option.AccessKey,
		SecretAccessKey: u.option.SecretAccessKey,
		BaseDir:         u.option.BaseDir,
		Version:         u.option.Version,
	}
	u.client = storageClient

	var (
		err     error
		exitCh  = make(chan *StorageResponse)
		summary []*restic.Snapshot
	)

	go u.run(ctx, exitCh)

	select {
	case e, ok := <-exitCh:
		if ok && e.Error != nil {
			err = e.Error
		}
		summary = e.SnapshotsSummary
	case <-ctx.Done():
		err = errors.Errorf("get snapshots timeout")
	}

	if err != nil {
		return err
	}

	if summary != nil && len(summary) > 0 {
		fmt.Printf("\n[RepoName]: %s\n", u.option.RepoName)
		fmt.Printf("-------------------------------------------------------\n")
		for _, snapshot := range summary {
			fmt.Printf("  Snapshot ID: %s\n", snapshot.ShortId)
			fmt.Printf("  Start Time: %s\n", snapshot.Summary.BackupStart)
			fmt.Printf("  Path: %v\n", snapshot.Paths)
			fmt.Printf("  Tags: %v\n", snapshot.Tags)
			fmt.Printf("  Files: %d\n", snapshot.Summary.TotalFilesProcessed)
			fmt.Printf("  Size: %s\n", util.FormatBytes(uint64(snapshot.Summary.TotalBytesProcessed)))
			fmt.Printf("\n")
		}
		fmt.Printf("-------------------------------------------------------\n\n")
	}

	return nil
}

func (u *Snapshots) run(ctx context.Context, exitCh chan<- *StorageResponse) {
	var olaresSpace = &OlaresSpace{
		RepoName:           u.client.RepoName,
		OlaresId:           u.client.OlaresId,
		StorageLocation:    u.client.StorageLocation,
		Endpoint:           u.client.Endpoint,
		AccessKey:          u.client.AccessKey,
		SecretAccessKey:    u.client.SecretAccessKey,
		CloudApiMirror:     u.client.CloudApiMirror,
		BackupsOperate:     OperateSnapshots,
		OlaresSpaceSession: new(OlaresSpaceSession),
		BaseDir:            u.client.BaseDir,
		Version:            u.client.Version,
	}

	if err := olaresSpace.GetAccount(); err != nil {
		exitCh <- &StorageResponse{Error: fmt.Errorf("get account error: %v", err)}
		return
	}

	if err := olaresSpace.EnterPassword(); err != nil {
		exitCh <- &StorageResponse{Error: err}
		return
	}

	if err := olaresSpace.GetToken(); err != nil { // snapshots
		exitCh <- &StorageResponse{Error: fmt.Errorf("get token error: %v", err)}
		return
	}

	// logger.Infof("get token, data: %s", util.ToJSON(olaresSpace))

	olaresSpace.SetRepoUrl()
	olaresSpace.SetEnv()

	logger.Debugf("get token, data: %s", util.Base64encode([]byte(util.ToJSON(olaresSpace))))

	r, err := restic.NewRestic(ctx, u.client.RepoName, u.client.OlaresId, olaresSpace.GetEnv(), &restic.Option{})
	if err != nil {
		exitCh <- &StorageResponse{Error: err}
		return
	}

	snapshotsSummary, err := r.GetSnapshots()
	if err != nil {
		exitCh <- &StorageResponse{Error: err}
		return
	}

	exitCh <- &StorageResponse{SnapshotsSummary: snapshotsSummary}
}
