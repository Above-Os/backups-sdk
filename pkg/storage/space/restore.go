package space

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/pkg/errors"
)

func (s *Space) Restore() (restoreSummary *restic.RestoreSummaryOutput, err error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	if err = s.getStsToken(); err != nil {
		return
	}

	storageInfo, err := s.FormatRepository()
	if err != nil {
		return
	}

	var restoreResult *restic.RestoreSummaryOutput

	for {
		var envs = s.GetEnv(storageInfo.Url)
		var opts = &restic.ResticOptions{
			RepoName:          s.RepoName,
			CloudName:         s.CloudName,
			RegionId:          s.RegionId,
			RepoEnvs:          envs,
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
		logger.Infof("space restore spanshot %s detail: %s", s.SnapshotId, utils.ToJSON(currentSnapshot))

		restoreSummary, err = r.Restore(s.SnapshotId, backupPath, s.Path)
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space restore download stopped, sts token expired, refresh and retring...")
				if err = s.refreshStsTokens(); err != nil {
					err = fmt.Errorf("space restore download sts token service refresh-token error: %v", err)
					return
				}
				continue
			default:
				return nil, errors.WithStack(err)
			}
		}

		logger.Infof("Restore successful, name: %s, result: %s", s.RepoName, utils.ToJSON(restoreResult))

		break
	}

	return
}
