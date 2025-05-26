package constants

const (
	DefaultBaseDir = ".olares"
	DefaultLogsDir = "logs"

	OlaresReleaseFile          = "/etc/olares/release"
	OlaresStorageDefaultPrefix = "olares-backups"

	ENV_OLARES_BASE_DIR = "OLARES_BASE_DIR"
	ENV_OLARES_VERSION  = "OLARES_VERSION"

	DefaultCloudApiUrl = "https://cloud-api.bttcdn.com"
	DefaultDownloadUrl = "https://dc3p1870nn3cj.cloudfront.net"

	StsTokenUrl        = "/v1/resource/stsToken/backup"
	StsTokenRefreshUrl = "/v1/resource/stsToken/backup/refresh"
	SendBackupUrl      = "/v1/resource/backup/save"
	SendSnapshotUrl    = "/v1/resource/snapshot/save"

	StorageS3Domain     = "amazonaws.com"
	StorageTencentDoman = "myqcloud.com"

	CloudAWSName        = "aws"
	CloudTencentName    = "tencentcloud"
	CloudFilesystemName = "filesystem"

	FullyBackup       string = "fully"
	IncrementalBackup string = "incremental"

	BackupCreate   = "create"
	BackupError    = "error"
	BackupComplete = "complete"

	StorageOperatorCli = "cli"
	StorageOperatorApp = "app"

	TraceId = "traceId"

	BackupTypeApp  = "app"
	BackupTypeFile = "file"
)
