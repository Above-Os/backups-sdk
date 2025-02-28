package tokens

import (
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space/tokens/space"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space/tokens/user"
)

type Tokens struct {
	UserToken  *user.UserToken   `json:"user_token"`
	SpaceToken *space.SpaceToken `json:"space_token"`
}

func (t *Tokens) IsTokensValid(olaresId, olaresName, repoName, repoRegion string) bool {
	if t.UserToken == nil || !t.UserToken.IsUserTokenValid(olaresId, olaresName) {
		return false
	}

	if t.SpaceToken == nil || !t.SpaceToken.IsSpaceTokenValid(repoName, repoRegion) {
		return false
	}

	return true
}

func (t *Tokens) GetTokens(olaresId, olaresName, repoLocation, repoRegion, cloudApiMirror string) error {
	err := t.UserToken.GetUserToken(olaresId, olaresName)
	if err != nil {
		return err
	}

	err = t.SpaceToken.GetSpaceToken(t.UserToken.OlaresKey, olaresId, olaresName, t.UserToken.SpaceUserAccessToken, repoLocation, repoRegion, cloudApiMirror)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tokens) RefreshTokens(olaresId, olaresName, repoLocation, repoRegion, cloudApiMirror string) error {
	if t.UserToken.IsSpaceUserAccessTokenExpired() {
		err := t.UserToken.GetUserToken(olaresId, olaresName)
		if err != nil {
			return err
		}
	}

	err := t.SpaceToken.RefreshSpaceToken(olaresName, cloudApiMirror)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tokens) GetSpaceEnv(repoName string, password string) *restic.ResticEnv {
	return t.SpaceToken.GetEnv(repoName, password)
}
