package space

import (
	"fmt"
	"path/filepath"

	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"github.com/pkg/errors"
)

type Space struct {
	RepoName        string
	SnapshotId      string
	RepoRegion      string
	OlaresId        string
	Password        string
	Path            string
	LimitUploadRate string
	CloudApiMirror  string
	BaseDir         string

	SpaceToken *SpaceToken
	UserToken  *UserToken
}

type StorageResponse struct {
	Summary          *restic.SummaryOutput
	RestoreSummary   *restic.RestoreSummaryOutput
	SnapshotsSummary []*restic.Snapshot
	Error            error
}

func (s *Space) GetSnapshotId() string {
	return s.SnapshotId
}

func (s *Space) GetPath() string {
	return s.Path
}

func (s *Space) GetRepoName() string {
	return s.RepoName
}

func (s *Space) GetLocation() common.Location {
	return common.LocationSpace
}

func (s *Space) GetLimitUploadRate() string {
	return s.LimitUploadRate
}

func (s *Space) GetEnv(repoName string) *restic.ResticEnv {
	s.RepoName = repoName
	repo, _ := s.FormatRepository()

	var envs = &restic.ResticEnv{
		AWS_ACCESS_KEY_ID:     s.SpaceToken.AccessKey,
		AWS_SECRET_ACCESS_KEY: s.SpaceToken.SecretKey,
		AWS_SESSION_TOKEN:     s.SpaceToken.SessionToken,
		RESTIC_REPOSITORY:     repo,
		RESTIC_PASSWORD:       s.Password,
	}

	return envs
}
func (s *Space) FormatRepository() (repository string, err error) {
	var repoPrefix = filepath.Join(s.SpaceToken.Prefix, "restic", s.RepoName)
	var domain = fmt.Sprintf("s3.%s.%s", s.SpaceToken.Region, common.AwsDomain)
	var repo = filepath.Join(domain, s.SpaceToken.Bucket, repoPrefix)
	repository = fmt.Sprintf("s3:%s", repo)
	return
}

func (s *Space) IsTokensValid(repoName, repoRegion string) bool {
	if s.UserToken == nil || !s.UserToken.IsUserTokenValid(s.OlaresId) {
		return false
	}

	if s.SpaceToken == nil || !s.SpaceToken.IsSpaceTokenValid(repoName, repoRegion) {
		return false
	}

	return true
}

func (s *Space) getTokens(repoLocation, repoRegion, cloudApiMirror string) error {
	if err := s.UserToken.GetUserToken(s.OlaresId); err != nil {
		return errors.WithStack(fmt.Errorf("get user token error: %v", err))
	}

	if err := s.SpaceToken.GetSpaceToken(s.UserToken.OlaresDid, s.UserToken.OlaresId, s.UserToken.SpaceUserAccessToken, repoLocation, repoRegion, cloudApiMirror); err != nil {
		return errors.WithStack(fmt.Errorf("get space token error: %v", err))
	}
	return nil
}

func (s *Space) refreshTokens(cloudApiMirror string) error {
	if err := s.UserToken.GetUserToken(s.OlaresId); err != nil {
		return errors.WithStack(fmt.Errorf("refresh user token error: %v", err))
	}

	if err := s.SpaceToken.RefreshSpaceToken(s.UserToken.OlaresId, cloudApiMirror); err != nil {
		return errors.WithStack(fmt.Errorf("refresh space token error: %v", err))
	}

	return nil
}
