package space

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
)

func (s *Space) Backup() error {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	var (
		err     error
		exitCh  = make(chan *StorageResponse)
		summary *restic.SummaryOutput
	)

	go s.runBackup(ctx, exitCh)

	select {
	case e, ok := <-exitCh:
		if ok && e.Error != nil {
			err = e.Error
		}
		summary = e.Summary
	case <-ctx.Done():
		err = errors.Errorf("space backup %q timed out", s.RepoName)
	}

	if err != nil {
		return err
	}

	if summary != nil {
		logger.Infof("space backup successful, data: %s", util.ToJSON(summary))
	}

	return nil
}

func (s *Space) runBackup(ctx context.Context, exitCh chan<- *StorageResponse) {
	var repoName = s.RepoName
	var olaresId = s.OlaresId
	var cloudApiMirror = s.CloudApiMirror
	var baseDir = s.BaseDir
	var password = s.Password
	var path = s.Path
	var repoLocation = "aws"
	var repoRegion = "us-east-1"

	svc, err := tokens.NewTokenService(olaresId)
	if err != nil {
		exitCh <- &StorageResponse{Error: fmt.Errorf("space backup token service error: %v", err)}
		return
	}

	var existsSpaceTokenCacheFile bool = true
	err = svc.InitSpaceTokenFromFile(baseDir)
	if err != nil {
		existsSpaceTokenCacheFile = false
	}

	var isTokensValid bool
	if existsSpaceTokenCacheFile {
		isTokensValid = svc.IsTokensValid(repoName, repoRegion)
	}

	if !isTokensValid {
		// todo write file
		logger.Infof("space backup tokens invalid, get new token")
		if err := svc.GetNewToken(repoLocation, repoRegion, cloudApiMirror); err != nil {
			exitCh <- &StorageResponse{Error: fmt.Errorf("space backup token service get-token error: %v", err)}
			return
		}
	}

	var summary *restic.SummaryOutput
	for {
		var resticEnv = svc.GetSpaceEnv(repoName, password)

		logger.Debugf("space backup env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

		r, err := restic.NewRestic(ctx, repoName, olaresId, resticEnv.ToMap(), &restic.Option{})
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}

		_, initRepo, err := r.Init()
		if err != nil {
			logger.Debugf("space backup init message: %s", err.Error())
			if err.Error() == restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error() {
				logger.Infof("space backup init stopped, token expired, refresh token and retring...")
				if err := svc.RefreshToken(repoLocation, repoRegion, cloudApiMirror); err != nil {
					exitCh <- &StorageResponse{Error: fmt.Errorf("space backup init token service refresh-token error: %v", err)}
					return
				}
				time.Sleep(2 * time.Second)
				continue
			} else {
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}

		if !initRepo {
			logger.Infof("space backup repair index, please wait...")
			if err := r.Repair(); err != nil {
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}

		logger.Infof("preparing to start space backup, repo: %s", repoName)

		summary, err = r.Backup(repoName, path, "")
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space backup upload stopped, token expired, refresh token and retring...")
				if err := svc.RefreshToken(repoLocation, repoRegion, cloudApiMirror); err != nil {
					exitCh <- &StorageResponse{Error: fmt.Errorf("space backup upload token service refresh-token error: %v", err)}
					return
				}
				r.NewContext()
				continue
			default:
				exitCh <- &StorageResponse{Error: err}
				return
			}
		}
		break
	}

	exitCh <- &StorageResponse{Summary: summary}
}
