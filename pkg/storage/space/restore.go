package space

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/pkg/errors"
)

func (s *Space) Restore(ctx context.Context, progressCallback func(percentDone float64)) (restoreSummary *restic.RestoreSummaryOutput, err error) {
	// ctx, cancel := context.WithCancel(context.TODO())
	// defer cancel()

	if err = s.getStsToken(ctx); err != nil {
		return
	}

	storageInfo, err := s.FormatRepository()
	if err != nil {
		return
	}

	var progressChan = make(chan float64, 100)
	defer close(progressChan)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				progressCallback(progress)
			}
		}
	}()

	for {
		var envs = s.GetEnv(storageInfo.Url)
		var opts = &restic.ResticOptions{
			RepoId:            s.RepoId,
			RepoName:          s.RepoName,
			CloudName:         s.CloudName,
			RegionId:          s.RegionId,
			RepoEnvs:          envs,
			Path:              s.Path,
			LimitDownloadRate: s.LimitDownloadRate,
		}

		logger.Debugf("space restore env vars: %s", utils.Base64encode([]byte(envs.String())))

		var r *restic.Restic
		r, err = restic.NewRestic(ctx, opts)
		if err != nil {
			break
		}

		var currentSnapshot *restic.Snapshot
		currentSnapshot, err = r.GetSnapshot(s.SnapshotId)
		if err != nil {
			return
		}

		var backupPath = currentSnapshot.Paths[0]

		for _, tag := range currentSnapshot.Tags {
			if tag == "content-type=files" {
				backupPath = ""
				break
			}
		}

		logger.Infof("space restore spanshot %s detail: %s", s.SnapshotId, utils.ToJSON(currentSnapshot))

		restoreSummary, err = r.Restore(s.SnapshotId, backupPath, s.Path, progressChan)
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space restore download stopped, sts token expired, refresh and retring...")
				if err = s.refreshStsTokens(ctx); err != nil {
					err = fmt.Errorf("space restore download sts token service refresh-token error: %v", err)
					return
				}
				continue
			default:
				return nil, errors.WithStack(err)
			}
		}

		logger.Infof("Restore successful, name: %s, result: %s", s.RepoName, utils.ToJSON(restoreSummary))

		break
	}

	return
}
