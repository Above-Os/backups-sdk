package storage

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"github.com/pkg/errors"
)

type SnapshotsProvider interface {
	Snapshots() error
}

type Snapshots struct {
	option SnapshotsOption
}

type SnapshotsOption struct {
	RepoName        string
	OlaresId        string
	BackupType      string
	Endpoint        string
	AccessKeyId     string
	SecretAccessKey string
	CloudApiMirror  string
}

func (u *Snapshots) Snapshots(opt SnapshotsOption) error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	u.option = opt

	var storageClient = &StorageClient{
		RepoName:        u.option.RepoName,
		OlaresId:        u.option.OlaresId,
		BackupType:      u.option.BackupType,
		Endpoint:        u.option.Endpoint,
		AccessKeyId:     u.option.AccessKeyId,
		SecretAccessKey: u.option.SecretAccessKey,
	}

	var (
		err     error
		exitCh  = make(chan *StorageResponse)
		summary []*restic.Snapshot
	)

	go storageClient.GetSnapshots(ctx, exitCh)

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
		fmt.Printf("RepoName: %s\n", u.option.RepoName)
		for _, snapshot := range summary {
			fmt.Printf("=======================================\n")
			fmt.Printf("snapshot id: %s\n", snapshot.ShortId)
			fmt.Printf("start: %s, end: %s\n", snapshot.Summary.BackupStart, snapshot.Summary.BackupEnd)
			fmt.Printf("paths: %v\n", snapshot.Paths)
			fmt.Printf("tags: %v\n", snapshot.Tags)
		}
	}

	return nil
}
