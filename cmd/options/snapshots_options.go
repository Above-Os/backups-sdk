package options

import (
	"bytetrade.io/web3os/backups-sdk/pkg/common"
	"github.com/spf13/cobra"
)

type SpaceSnapshotsOption struct {
	RepoName       string
	OlaresId       string
	CloudApiMirror string
	BaseDir        string
}

func NewSnapshotsSpaceOption() *SpaceSnapshotsOption {
	return &SpaceSnapshotsOption{}
}

func (o *SpaceSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.CloudApiMirror, "cloud-api-mirror", "", "", "Cloud API mirror")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ s3
type S3SnapshotsOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	OlaresId        string
	BaseDir         string
}

func NewSnapshotsS3Option() *S3SnapshotsOption {
	return &S3SnapshotsOption{}
}

func (o *S3SnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for S3, for example https://{bucket}.{region}.amazonaws.com/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for S3")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for S3")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ cos
type CosSnapshotsOption struct {
	RepoName        string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	OlaresId        string
	BaseDir         string
}

func NewSnapshotsCosOption() *CosSnapshotsOption {
	return &CosSnapshotsOption{}
}

func (o *CosSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "Endpoint for Tencent COS, for example https://cos.{region}.myqcloud.com/{bucket}/{prefix}")
	cmd.Flags().StringVarP(&o.AccessKey, "access-key", "", "", "Access Key for Tencent COS")
	cmd.Flags().StringVarP(&o.SecretAccessKey, "secret-access-key", "", "", "Secret Access Key for Tencent COS")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}

// ~ filesystem
type FilesystemSnapshotsOption struct {
	RepoName string
	Endpoint string
	OlaresId string
	BaseDir  string
}

func NewSnapshotsFilesystemOption() *FilesystemSnapshotsOption {
	return &FilesystemSnapshotsOption{}
}

func (o *FilesystemSnapshotsOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.RepoName, "repo-name", "", "", "Backup repo name")
	cmd.Flags().StringVarP(&o.Endpoint, "endpoint", "", "", "The endpoint of the filesystem is the local computer directory where the backup will be stored")
	cmd.Flags().StringVarP(&o.OlaresId, "olares-id", "", "", "Olares ID")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "", "", "Set Olares package base dir, defaults to $HOME/"+common.DefaultBaseDir)
}
