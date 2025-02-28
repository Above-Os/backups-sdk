package space

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space/tokens"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
)

func (s *Space) Snapshots() error {
	var repoName = s.RepoName
	var olaresId = s.OlaresId
	var cloudApiMirror = s.CloudApiMirror
	var baseDir = s.BaseDir
	var password = s.Password
	var repoLocation = "aws"
	var repoRegion = "us-east-1"

	svc, err := tokens.NewTokenService(olaresId)
	if err != nil {
		return errors.WithStack(fmt.Errorf("space token service error: %v", err))
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
		logger.Infof("space snapshots tokens invalid, get new token")
		if err := svc.GetNewToken(repoLocation, repoRegion, cloudApiMirror); err != nil {
			return errors.WithStack(fmt.Errorf("space snapshots token service get-token error: %v", err))
		}
	}

	var resticEnv = svc.GetSpaceEnv(repoName, password)
	logger.Debugf("space snapshots env vars: %s", util.Base64encode([]byte(resticEnv.ToString())))

	r, err := restic.NewRestic(context.Background(), repoName, "", resticEnv.ToMap(), nil)
	if err != nil {
		return err
	}

	snapshots, err := r.GetSnapshots()
	if err != nil {
		return err
	}
	snapshots.PrintTable()

	return nil
}
