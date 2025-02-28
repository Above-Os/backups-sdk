package space

import "bytetrade.io/web3os/backups-sdk/pkg/response"

type CloudStorageAccountResponse struct {
	response.Header
	Data *OlaresSpaceSession `json:"data"`
}

type OlaresSpaceSession struct {
	Cloud          string `json:"cloud"`
	Bucket         string `json:"bucket"`
	Token          string `json:"st"`
	Prefix         string `json:"prefix"`
	Secret         string `json:"sk"`
	Key            string `json:"ak"`
	Expiration     string `json:"expiration"`
	Region         string `json:"region"`
	ResticRepo     string `json:"restic_repo"`
	ResticPassword string `json:"-"`
}
