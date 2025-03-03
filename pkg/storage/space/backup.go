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
		// todo print info
		logger.Infof("space backup successful, data: %s", util.ToJSON(summary))
	}

	return nil
}

func (s *Space) runBackup(ctx context.Context, exitCh chan<- *StorageResponse) {
	var repoName = s.RepoName
	var cloudApiMirror = s.CloudApiMirror
	var baseDir = s.BaseDir
	var path = s.Path
	var repoLocation = "aws"
	var repoRegion = "us-east-1"
	_ = baseDir

	// get user token and space aws session-token
	if err := s.getTokens(repoLocation, repoRegion, cloudApiMirror); err != nil {
		exitCh <- &StorageResponse{Error: err}
		return
	}

	var summary *restic.SummaryOutput
	for {
		envs := s.GetEnv(repoName)

		logger.Debugf("space backup env vars: %s", util.Base64encode([]byte(envs.ToString())))

		r, err := restic.NewRestic(ctx, repoName, envs, &restic.Option{LimitUploadRate: s.LimitUploadRate})
		if err != nil {
			exitCh <- &StorageResponse{Error: err}
			return
		}

		_, initRepo, err := r.Init()
		if err != nil {
			logger.Debugf("space backup init message: %s", err.Error())
			if err.Error() == restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error() {
				logger.Infof("space backup init stopped, token expired, refresh token and retring...")
				if err := s.refreshTokens(cloudApiMirror); err != nil {
					exitCh <- &StorageResponse{Error: err}
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

		summary, err = r.Backup(path, "")
		if err != nil {
			switch err.Error() {
			case restic.ERROR_MESSAGE_TOKEN_EXPIRED.Error():
				logger.Infof("space backup upload stopped, token expired, refresh token and retring...")
				if err := s.refreshTokens(cloudApiMirror); err != nil {
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
