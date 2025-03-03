package common

const (
	DefaultBaseDir       = ".olares"
	DefaultLogDir        = "logs"
	DefaultConsoleLogDir = "backups"
	DefaultCloudApiUrl   = "https://cloud-api.bttcdn.com"

	StorageLocationOlaresAWS  = "aws"
	StorageLocationS3         = "s3"
	StorageLocationCos        = "cos"
	StorageLocationFilesystem = "local"

	DefaultBackupOlaresRegion = "us-east-1"

	AwsDomain     = "amazonaws.com"
	TencentDomain = "myqcloud.com"
	AliyunDomain  = "aliyuncs.com"
)

type Location string

const (
	LocationSpace      Location = "space"
	LocationS3         Location = "s3"
	LocationCos        Location = "cos"
	LocationFileSystem Location = "filesystem"
)

type Operation string

const (
	Backup    Operation = "backup"
	Restore   Operation = "restore"
	Snapshots Operation = "snapshots"
)
