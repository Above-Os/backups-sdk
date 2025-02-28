package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/client"
	"bytetrade.io/web3os/backups-sdk/pkg/restic"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space/tokens/space"
	"bytetrade.io/web3os/backups-sdk/pkg/storage/space/tokens/user"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

type TokenHandler interface {
	InitSpaceTokenFromFile(baseDir string) error
	IsTokensValid(repoName, repoRegion string) bool
	GetNewToken(repoLocation, repoRegion, cloudApiMirror string) error
	RefreshToken(repoLocation, repoRegion, cloudApiMirror string) error

	GetSpaceEnv(repoName, password string) *restic.ResticEnv
}

type TokensService struct {
	tokens     *Tokens
	olaresId   string
	olaresName string
}

func NewTokenService(olaresId string) (TokenHandler, error) {
	var svc = new(TokensService)

	olaresName, err := svc.getOlaresName(olaresId)
	if err != nil {
		return nil, err
	}

	svc.olaresId = olaresId
	svc.olaresName = olaresName
	svc.tokens = &Tokens{
		UserToken:  new(user.UserToken),
		SpaceToken: new(space.SpaceToken),
	}

	return svc, nil
}

func (s *TokensService) InitSpaceTokenFromFile(baseDir string) error {
	var f = filepath.Join(baseDir, ".backups")
	content, err := util.ReadFile(f)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &s.tokens); err != nil {
		return err
	}

	if s.tokens == nil || s.tokens.UserToken == nil || s.tokens.SpaceToken == nil {
		logger.Debugf("tokens from file is invalid, data: %s", string(content))
		return errors.New("token files invalid")
	}

	return nil
}

func (s *TokensService) IsTokensValid(repoName, repoRegion string) bool {
	return s.tokens.IsTokensValid(s.olaresId, s.olaresName, repoName, repoRegion)
}

func (s *TokensService) GetNewToken(repoLocation, repoRegion, cloudApiMirror string) error {
	return s.tokens.GetTokens(s.olaresId, s.olaresName, repoLocation, repoRegion, cloudApiMirror)
}

func (s *TokensService) RefreshToken(repoLocation, repoRegion, cloudApiMirror string) error {
	return s.tokens.RefreshTokens(s.olaresId, s.olaresName, repoLocation, repoRegion, cloudApiMirror)
}

func (s *TokensService) GetSpaceEnv(repoName, password string) *restic.ResticEnv {
	return s.tokens.GetSpaceEnv(repoName, password)
}

// todo
func (t *TokensService) write(baseDir string) error {
	c, err := t.parse()
	if err != nil {
		return err
	}

	var f = filepath.Join(baseDir, ".backups")
	return util.WriteFile(f, c, 0644)
}

// todo
func (t *TokensService) parse() ([]byte, error) {
	// var data = make(map[string]interface{})
	// data["user_token"] = t.token.User
	// data["storage_token"] = t.token.Space

	// content, err := json.Marshal(t)
	// if err != nil {
	// 	return nil, err
	// }

	// return content, nil

	return nil, nil
}

func (t *TokensService) getOlaresName(olaresId string) (olaresName string, err error) {
	factory, err := client.NewFactory()
	if err != nil {
		return
	}

	dynamicClient, err := factory.DynamicClient()
	if err != nil {
		return
	}

	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    2,
	}
	var gvr = schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}

	if err = retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// todo unittest
		unstructuredUser, err := dynamicClient.Resource(gvr).Get(ctx, olaresId, metav1.GetOptions{})
		if err != nil {
			return errors.WithStack(fmt.Errorf("get space user res error: %v", err))
		}
		obj := unstructuredUser.UnstructuredContent()
		olaresName, _, err = unstructured.NestedString(obj, "spec", "email")
		if err != nil {
			return errors.WithStack(fmt.Errorf("get space user name error: %v", err))
		}
		return nil
	}); err != nil {
		return
	}

	return
}
