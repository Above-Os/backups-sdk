package common

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	DefaultCloudApiUrl = "https://cloud-api.bttcdn.com"

	DefaultStorageLocation    = "olares-space"
	StorageLocationOlaresAWS  = "aws"
	StorageLocationS3         = "s3"
	StorageLocationCos        = "cos"
	StorageLocationFilesystem = "local"

	DefaultBackupOlaresRegion = "us-east-1"

	AwsDomain     = "amazonaws.com"
	TencentDomain = "myqcloud.com"
	AliyunDomain  = "aliyuncs.com"
)

var TerminusGVR = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "terminus",
}

var UsersGVR = schema.GroupVersionResource{
	Group:    "iam.kubesphere.io",
	Version:  "v1alpha2",
	Resource: "users",
}
