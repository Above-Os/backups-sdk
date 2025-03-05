package space

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
)

func (s *Space) Restore() error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	var (
		err     error
		exitCh  = make(chan *StorageResponse)
		summary *restic.RestoreSummaryOutput
	)

	go s.runRestore(ctx, exitCh)

	select {
	case e, ok := <-exitCh:
		if ok && e.Error != nil {
			err = e.Error
		}
		summary = e.RestoreSummary
	case <-ctx.Done():
		err = errors.Errorf("space restore %s time out", s.RepoName)
	}

	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("restore space successful, data: %s", util.ToJSON(summary))
	}

	return nil
}

func (s *Space) runRestore(ctx context.Context, exitCh chan<- *StorageResponse) {
	var repoName = s.RepoName

	// get space sts token
	if err := s.getStsToken(s.CloudName, s.RegionId); err != nil {
		exitCh <- &StorageResponse{Error: err}
		return
	}

	var summary *restic.RestoreSummaryOutput
	for {
		var envs = s.GetEnv(repoName)
		var opts = &restic.ResticOptions{
			RepoName:        s.RepoName,
			RepoEnvs:        envs,
			LimitUploadRate: s.LimitUploadRate,
		}

		logger.Debugf("space restore env vars: %s", util.Base64encode([]byte(envs.String())))

		r, err := restic.NewRestic(ctx, opts)
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}

		snapshotSummary, err := r.GetSnapshot(s.SnapshotId)
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}
		var uploadPath = snapshotSummary.Paths[0]
		logger.Infof("space restore spanshot %s detail: %s", s.SnapshotId, util.ToJSON(snapshotSummary))

		summary, err = r.Restore(s.SnapshotId, uploadPath, s.Path)
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space restore download stopped, sts token expired, refresh and retring...")
				if err := s.refreshStsTokens(); err != nil {
					exitCh <- &StorageResponse{Error: fmt.Errorf("space restore download sts token service refresh-token error: %v", err)}
					return
				}
				continue
			default:
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}
		break

	}
	exitCh <- &StorageResponse{RestoreSummary: summary}
}
