package space

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space/tokens"
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
	var snapshotId = s.SnapshotId
	var olaresId = s.OlaresId
	var cloudApiMirror = s.CloudApiMirror
	var baseDir = s.BaseDir
	var password = s.Password
	var path = s.Path
	var repoLocation = "aws"
	var repoRegion = "us-east-1"

	svc, err := tokens.NewTokenService(olaresId)
	if err != nil {
		exitCh <- &StorageResponse{Error: fmt.Errorf("space token service error: %v", err)}
		return
	}

	var existsSpaceTokenCacheFile bool = true
	err = svc.InitSpaceTokenFromFile(baseDir)
	if err != nil {
		existsSpaceTokenCacheFile = false
	}

	var isTokenValid bool
	if existsSpaceTokenCacheFile {
		isTokenValid = svc.IsTokensValid(repoName, repoRegion)
	}

	if !isTokenValid {
		// todo write file
		if err := svc.GetNewToken(repoLocation, repoRegion, cloudApiMirror); err != nil {
			exitCh <- &StorageResponse{Error: fmt.Errorf("space restore token service get-token error: %v", err)}
			return
		}
	}

	var summary *restic.RestoreSummaryOutput
	for {
		var resticEnv = svc.GetSpaceEnv(repoName, password)

		logger.Debugf("space restore env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

		r, err := restic.NewRestic(ctx, repoName, olaresId, resticEnv.ToMap(), &restic.Option{})
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}

		snapshotSummary, err := r.GetSnapshot(snapshotId)
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}
		var uploadPath = snapshotSummary.Paths[0]
		logger.Infof("space restore spanshot %s detail: %s", snapshotId, util.ToJSON(snapshotSummary))

		summary, err = r.Restore(snapshotId, uploadPath, path)
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space restore download stopped, token expired, refresh token and retring...")
				if err := svc.RefreshToken(repoLocation, repoRegion, cloudApiMirror); err != nil { // restore
					exitCh <- &StorageResponse{Error: fmt.Errorf("space restore download token service refresh-token error: %v", err)}
					return
				}
				r.NewContext()
				time.Sleep(2 * time.Second)
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
