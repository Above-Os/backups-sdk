package vars

import "bytetrade.io/web3os/backups-sdk/pkg/response"

type AccountResponse struct {
	response.Header
	Data *AccountResponseData `json:"data,omitempty"`
}

type AccountResponseRawData struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	// ExpiresIn    int64  `json:"expires_in"`
	ExpiresAt int64  `json:"expires_at"`
	UserId    string `json:"userid"`
	Available bool   `json:"available"`
	CreateAt  int64  `json:"create_at"`
}

type AccountResponseData struct {
	Name     string                  `json:"name"`
	Type     string                  `json:"type"`
	RawData  *AccountResponseRawData `json:"raw_data"`
	CloudUrl string                  `json:"cloudUrl"`
}

type AccountValue struct {
	Email   string `json:"email"`
	Userid  string `json:"userid"`
	Token   string `json:"token"`
	Expired any    `json:"expired"`
}
