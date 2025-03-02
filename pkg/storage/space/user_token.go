package space

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/client"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/util/net"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	HEADER_NONCE = "Terminus-Nonce"
)

type UserToken struct {
	OlaresId             string `json:"olares_id"`
	OlaresName           string `json:"olares_name"`
	OlaresKey            string `json:"olares_key"`
	SpaceUserAccessToken string `json:"space_user_access_token"`
	ExpiresAt            int64  `json:"expires_at"`
	CreateAt             int64  `json:"create_at"`
}

func (u *UserToken) IsUserTokenValid(olaresId, olaresName string) bool {
	return u.isSpaceUserNameMatched(olaresId, olaresName) && !u.isSpaceUserAccessTokenExpired()
}

func (u *UserToken) IsSpaceUserAccessTokenExpired() bool {
	return u.isSpaceUserAccessTokenExpired()
}

func (u *UserToken) GetUserToken(olaresId, olaresName string) error {
	logger.Infof("get user %s token", olaresName)

	podIp, err := u.getPodIp(olaresId)
	if err != nil {
		return err
	}

	if err = u.getSpaceUserAccessToken(olaresId, olaresName, podIp); err != nil {
		return err
	}

	return nil
}

//
//
//

func (u *UserToken) isSpaceUserNameMatched(olaresId, olaresName string) bool {
	if u.OlaresId == "" || u.OlaresName == "" || u.OlaresKey == "" || u.OlaresId != olaresId || u.OlaresName != olaresName {
		return false
	}
	return true
}

func (u *UserToken) isSpaceUserAccessTokenExpired() bool {
	expiresTime, expired := util.IsTimestampAboutToExpire(u.ExpiresAt)
	logger.Infof("user access token expires at %s", expiresTime.String())
	return expired
}

func (u *UserToken) getAppKey() (string, error) {
	factory, err := client.NewFactory()
	if err != nil {
		return "", errors.WithStack(err)
	}

	kubeClient, err := factory.KubeClient()
	if err != nil {
		return "", errors.WithStack(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	secret, err := kubeClient.CoreV1().Secrets("os-system").Get(ctx, "app-key", metav1.GetOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}
	if secret == nil || secret.Data == nil || len(secret.Data) == 0 {
		return "", fmt.Errorf("secret not found")
	}

	key, ok := secret.Data["random-key"]
	if !ok {
		return "", fmt.Errorf("app key not found")
	}

	return string(key), nil
}

func (u *UserToken) getPodIp(olaresId string) (string, error) {
	factory, err := client.NewFactory()
	if err != nil {
		return "", errors.WithStack(err)
	}

	kubeClient, err := factory.KubeClient()
	if err != nil {
		return "", errors.WithStack(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pods, err := kubeClient.CoreV1().Pods(fmt.Sprintf("user-system-%s", olaresId)).List(ctx, metav1.ListOptions{
		LabelSelector: "app=systemserver",
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	if pods == nil || pods.Items == nil || len(pods.Items) == 0 {
		return "", fmt.Errorf("system server pod not found")
	}

	pod := pods.Items[0]
	podIp := pod.Status.PodIP
	if podIp == "" {
		return "", fmt.Errorf("system server pod ip invalid")
	}

	return podIp, nil
}

func (u *UserToken) getSpaceUserAccessToken(olaresId, olaresName string, podIp string) error {
	var headers, err = u.getRequestUserHeader()
	if err != nil {
		return err
	}
	var data = u.getRequestUserData(olaresName)
	var url = u.getRequestUserUrl(podIp)

	result, err := net.Post[AccountResponse](url, headers, data, true, false)
	if err != nil {
		return err
	}
	accountResp := result
	if accountResp.Code == 1 && accountResp.Message == "" {
		err = errors.WithStack(fmt.Errorf("\nOlaresSpace is not enabled. Please go to the Settings - Integration page in the LarePass App to add OlaresSpace\n"))
		return err
	} else if accountResp.Code != 0 {
		err = errors.WithStack(fmt.Errorf("request account settings api response error, status: %d, message: %s", accountResp.Code, accountResp.Message))
		return err
	}

	if accountResp.Data == nil || accountResp.Data.RawData == nil {
		err = errors.WithStack(fmt.Errorf("request account settings api response data is nil, status: %d, message: %s", accountResp.Code, accountResp.Message))
		return err
	}

	if accountResp.Data.RawData.UserId == "" || accountResp.Data.RawData.AccessToken == "" {
		err = errors.WithStack(fmt.Errorf("space user access token is empty"))
		return err
	}

	expiresAt, expired := util.IsTimestampAboutToExpire(accountResp.Data.RawData.ExpiresAt)
	logger.Infof("space user access token expires at %s", expiresAt.String())
	if expired {
		err = errors.WithStack(fmt.Errorf("space user access token is invalid, expires at %s", expiresAt.String()))
		return err
	}

	u.OlaresId = olaresId
	u.OlaresName = olaresName
	u.OlaresKey = accountResp.Data.RawData.UserId
	u.SpaceUserAccessToken = accountResp.Data.RawData.AccessToken
	u.ExpiresAt = accountResp.Data.RawData.ExpiresAt
	u.CreateAt = accountResp.Data.RawData.CreateAt

	return nil
}

func (u *UserToken) generateNonce(randomKey string) (string, error) {
	if randomKey == "" {
		randomKey = os.Getenv("APP_RANDOM_KEY")
	}
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	cipherText, err := util.AesEncrypt([]byte(timestamp), []byte(randomKey))
	if err != nil {
		return "", err
	}
	b64CipherText := base64.StdEncoding.EncodeToString(cipherText)
	headerNonce := "appservice:" + b64CipherText
	return headerNonce, nil
}

func (u *UserToken) getRequestUserUrl(ip string) string {
	return fmt.Sprintf("http://%s/legacy/v1alpha1/service.settings/v1/api/account/retrieve", ip)
}

func (u *UserToken) getRequestUserHeader() (map[string]string, error) {
	appKey, err := u.getAppKey()
	if err != nil {
		return nil, err
	}

	headerNonce, err := u.generateNonce(appKey)
	if err != nil {
		return nil, fmt.Errorf("generate nonce error: %v", err)
	}
	var headers = map[string]string{
		restful.HEADER_ContentType: restful.MIME_JSON,
		HEADER_NONCE:               headerNonce,
	}
	return headers, nil
}

func (u *UserToken) getRequestUserData(olaresName string) map[string]interface{} {
	var data = make(map[string]interface{})
	data["name"] = fmt.Sprintf("integration-account:space:%s", olaresName)
	return data
}
