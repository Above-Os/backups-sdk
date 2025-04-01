package model

type StorageInfo struct {
	Location  string `json:"location"`
	Url       string `json:"url"`
	CloudName string `json:"cloud_name"`
	RegionId  string `json:"region_id"`
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix"`
}
