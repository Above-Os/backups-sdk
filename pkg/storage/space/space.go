package space

import (
	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"fmt"
	"path/filepath"
)

type Space struct {
	RepoName       string
	SnapshotId     string
	RepoRegion     string
	OlaresId       string
	Password       string
	Path           string
	CloudApiMirror string
	BaseDir        string

	SpaceToken *SpaceToken
	UserToken  *UserToken
}

type StorageResponse struct {
	Summary          *restic.SummaryOutput
	RestoreSummary   *restic.RestoreSummaryOutput
	SnapshotsSummary []*restic.Snapshot
	Error            error
}

func (s *Space) GetEnv(repoName string) *restic.ResticEnv {
	s.RepoName = repoName
	repo, _ := s.FormatRepository()

	var envs = &restic.ResticEnv{
		AWS_ACCESS_KEY_ID:     s.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.SecretKey,
		AWS_SESSION_TOKEN:     s.SessionToken,
		RESTIC_REPOSITORY:     repo,
		RESTIC_PASSWORD:       s.Password,
	}

	return envs
}
func (s *Space) FormatRepository() (repository string, err error) {
	var repoPrefix = filepath.Join(s.Prefix, "restic", s.RepoName)
	var domain = fmt.Sprintf("s3.%s.%s", s.Region, common.AwsDomain)
	var repo = filepath.Join(domain, s.Bucket, repoPrefix)
	repository = fmt.Sprintf("s3:%s", repo)
	return
}

func (s *Space) IsTokensValid(repoName, repoRegion string) bool {
	if s.UserToken == nil || !s.UserToken.IsUserTokenValid(s.OlaresId, olaresName) {
		return false
	}

	if s.SpaceToken == nil || !s.SpaceToken.IsSpaceTokenValid(repoName, repoRegion) {
		return false
	}

	return true
}

func (s *Space) GetNewToken(repoLocation, repoRegion, cloudApiMirror string) error {
	err := s.UserToken.GetUserToken(s.OlaresId, olaresName)
	if err != nil {
		return err
	}

	err = s.SpaceToken.GetSpaceToken(s.UserToken.OlaresKey, s.OlaresId, olaresName, s.UserToken.SpaceUserAccessToken, repoLocation, repoRegion, cloudApiMirror)
	if err != nil {
		return err
	}

	return nil
}

func (s *Space) RefreshToken(repoLocation, repoRegion, cloudApiMirror string) error {
	if s.UserToken.IsSpaceUserAccessTokenExpired() {
		err := s.UserToken.GetUserToken(s.OlaresId, olaresName)
		if err != nil {
			return err
		}
	}

	err := s.SpaceToken.RefreshSpaceToken(olaresName, cloudApiMirror)
	if err != nil {
		return err
	}

	return nil
}
