package restic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"olares.com/backups-sdk/pkg/constants"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/utils"
)

type RESTIC_ERROR_MESSAGE string

const (
	SUCCESS_MESSAGE_REPAIR_INDEX                     RESTIC_ERROR_MESSAGE = "adding pack file to index"
	ERROR_MESSAGE_UNABLE_TO_OPEN_REPOSITORY          RESTIC_ERROR_MESSAGE = "unable to open repository at"
	ERROR_MESSAGE_TOKEN_EXPIRED                      RESTIC_ERROR_MESSAGE = "The provided token has expired"
	ERROR_MESSAGE_COS_TOKEN_EXPIRED                  RESTIC_ERROR_MESSAGE = "The Access Key Id you provided does not exist in our records"
	ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE         RESTIC_ERROR_MESSAGE = "unable to open config file: Stat: 400 Bad Request"
	ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE_MESSAGE RESTIC_ERROR_MESSAGE = "storage address is incorrect, unable to locate the configuration file"
	ERROR_MESSAGE_CONFIG_INVALID                     RESTIC_ERROR_MESSAGE = "config invalid, please chek repository or authorization config"
	ERROR_MESSAGE_LOCKED                             RESTIC_ERROR_MESSAGE = "repository is already locked by"
	ERROR_MESSAGE_ALREADY_INITIALIZED                RESTIC_ERROR_MESSAGE = "repository master key and config already initialized"
	ERROR_MESSAGE_SNAPSHOT_NOT_FOUND                 RESTIC_ERROR_MESSAGE = "no matching ID found for prefix"
	ERROR_MESSAGE_CONFIG_FILE_ALREADY_EXISTS         RESTIC_ERROR_MESSAGE = "config file already exists"
	ERROR_MESSAGE_WRONG_PASSWORD_OR_NO_KEY_FOUND     RESTIC_ERROR_MESSAGE = "wrong password or no key found"
	ERROR_MESSAGE_WRONG_PASSWORD                     RESTIC_ERROR_MESSAGE = "Wrong backup password."
	ERROR_MESSAGE_REPOSITORY_DOES_NOT_EXIST          RESTIC_ERROR_MESSAGE = "repository does not exist: unable to open config file"
	ERROR_MESSAGE_REPOSITORY_DOES_NOT_EXIST_MESSAGE  RESTIC_ERROR_MESSAGE = "Unable to locate the configuration file. Please provide the correct storage path."
	ERROR_MESSAGE_BACKUP_CANCELED                    RESTIC_ERROR_MESSAGE = "backup canceled"
	ERROR_MESSAGE_RESTORE_CANCELED                   RESTIC_ERROR_MESSAGE = "restore canceled"
	ERROR_MESSAGE_FILES_NOT_FOUND                    RESTIC_ERROR_MESSAGE = "does not match any files"
	ERROR_MESSAGE_SERVER_MISBEHAVING                 RESTIC_ERROR_MESSAGE = "server misbehaving"
	ERROR_MESSAGE_SERVER_MISBEHAVING_MESSAGE         RESTIC_ERROR_MESSAGE = "Network connection error."
	ERROR_MESSAGE_REPOSITORY_BE_DAMAGED              RESTIC_ERROR_MESSAGE = "the repository could be damaged"
	ERROR_MESSAGE_REPOSITORY_BE_DAMAGED_MESSAGE      RESTIC_ERROR_MESSAGE = "The repository might be corrupted. Please retry the snapshot task."
	ERROR_MESSAGE_COS_ACCOUNT_ARREARS                RESTIC_ERROR_MESSAGE = "Due to your account is arrears, it is unavailable until you recharge."
	ERROR_MESSAGE_COS_ACCOUNT_ARREARS_MESSAGE        RESTIC_ERROR_MESSAGE = "Your cloud storage account is overdue and the service is temporarily unavailable. Please recharge to continue."
	ERROR_MESSAGE_RESOURCE_TEMPORARILY_UNAVAILABLE   RESTIC_ERROR_MESSAGE = "resource temporarily unavailable"
	ERROR_MESSAGE_NO_SUCH_DEVICE                     RESTIC_ERROR_MESSAGE = "no such device"
	ERROR_MESSAGE_NO_SUCH_FILE_OR_DIRECTORY          RESTIC_ERROR_MESSAGE = "no such file or directory"
	ERROR_MESSAGE_NO_SUCH_DEVICE_MESSAGE             RESTIC_ERROR_MESSAGE = "No storage device found."
	ERROR_MESSAGE_HOST_IS_DOWN                       RESTIC_ERROR_MESSAGE = "host is down"
	ERROR_MESSAGE_HOST_IS_DOWN_MESSAGE               RESTIC_ERROR_MESSAGE = "SMB host down."
	ERROR_MESSAGE_NO_SPACE_LEFT_ON_DEVICE            RESTIC_ERROR_MESSAGE = "no space left on device"
	ERROR_MESSAGE_NO_SPACE_LEFT_ON_DEVICE_MESSAGE    RESTIC_ERROR_MESSAGE = "Insufficient storage."
	ERROR_MESSAGE_ACCESS_DENIED                      RESTIC_ERROR_MESSAGE = "Access Denied"
	ERROR_MESSAGE_ACCESS_DENIED_MESSAGE              RESTIC_ERROR_MESSAGE = "Access denied. Please provide the correct access key."
)

const (
	MESSAGE_TOKEN_EXPIRED                  = "[INFO] sts token expired"
	MESSAGE_REPOSITORY_ALREADY_INITIALIZED = "[INFO] repository already initialized"
)

const (
	PARAM_JSON_OUTPUT  = "--json"
	PARAM_INSECURE_TLS = "--insecure-tls"
)

func (e RESTIC_ERROR_MESSAGE) Error() string {
	return string(e)
}

func (e RESTIC_ERROR_MESSAGE) ToLower() string {
	return strings.ToLower(string(e))
}

const (
	tolerance = 1e-9

	PRINT_START_MESSAGE    = "[Upload] start, files: %d, size: %s"
	PRINT_PROGRESS_MESSAGE = "[Upload] progress %s, files: %d/%d, size: %s/%s, current: %v"
	PRINT_FINISH_MESSAGE   = "[Upload] finished, files: %d, size: %s, please waiting..."

	PRINT_RESTORE_START_MESSAGE    = "[Download] start, files: %d, size: %s, please waiting..."
	PRINT_RESTORE_PROGRESS_MESSAGE = "[Download] progress %s, files: %d/%d, size: %s/%s"
	PRINT_RESTORE_ITEM             = "[Download] restored file: %s, size: %s"
	PRINT_RESTORE_FINISH_MESSAGE   = "[Download] snapshot %s finished, total files: %d, restored files: %d, total size: %s, restored size: %s, please waiting..."
	PRINT_SUCCESS_MESSAGE          = ""
)

type ResticOptions struct {
	RepoId            string
	RepoName          string
	RepoSuffix        string
	CloudName         string
	RegionId          string
	SnapshotId        string
	Path              string
	Files             []string
	FilesPrefixPath   string
	Metadata          string
	LimitDownloadRate string
	LimitUploadRate   string
	DryRun            bool

	Operator                 string
	BackupType               string
	BackupAppTypeName        string
	BackupFileTypeSourcePath string
	RepoEnvs                 *ResticEnvs
}

func (o *ResticOptions) SetLimitUploadRate() string {
	var defaultUploadRate = "--limit-upload=0"
	if o.LimitUploadRate == "" {
		return defaultUploadRate
	}

	res, err := strconv.ParseInt(o.LimitUploadRate, 10, 64)
	if err != nil {
		return defaultUploadRate
	}

	return fmt.Sprintf("--limit-upload=%d", res)
}

func (o *ResticOptions) SetLimitDownloadRate() string {
	var defaultDownloadRate = "--limit-download=0"
	if o.LimitDownloadRate == "" {
		return defaultDownloadRate
	}

	res, err := strconv.ParseInt(o.LimitDownloadRate, 10, 64)
	if err != nil {
		return defaultDownloadRate
	}

	return fmt.Sprintf("--limit-download=%d", res)
}

type Restic struct {
	ctx    context.Context
	cancel context.CancelFunc
	dir    string
	args   []string
	opt    *ResticOptions
}

func NewRestic(ctx context.Context, opt *ResticOptions) (*Restic, error) {
	var commandPath, err = utils.Lookup("restic")
	if err != nil {
		return nil, fmt.Errorf("restic not found")
	}
	var ctxRestic, cancel = context.WithCancel(ctx)
	return &Restic{
		ctx:    ctxRestic,
		cancel: cancel,
		dir:    commandPath,
		opt:    opt,
	}, nil
}

func (r *Restic) Init() (string, error) {
	r.addCommand([]string{"init", "-v=3", PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS}).addExtended()

	cmd := exec.CommandContext(r.ctx, r.dir, r.args...)
	cmd.Env = append(cmd.Env, r.opt.RepoEnvs.Slice()...)
	logger.Infof("[Cmd] %s", cmd.String())
	var outerr RESTIC_ERROR_MESSAGE
	output, err := cmd.CombinedOutput()
	if err != nil {
		var errmsg = string(output)
		outerr, _ = r.formatErrorMessage(errmsg)
	}

	if outerr.Error() != "" {
		return "", errors.New(outerr.Error())
	}

	return string(output), nil
}

func (r *Restic) Rollback() error {
	backoff := wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    10,
	}
	var errorMsg RESTIC_ERROR_MESSAGE
	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		res, _ := r.prune()
		if strings.Contains(res, ERROR_MESSAGE_LOCKED.Error()) {
			r.Unlock()
			return fmt.Errorf("retry")
		}
		errorMsg, _ = r.formatErrorMessage(res)
		return nil
	}); err != nil {
		return err
	}
	if errorMsg.Error() != "" {
		return errorMsg
	}
	return nil
}

func (r *Restic) prune() (string, error) {
	var getCtx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	r.addCommand([]string{"prune", PARAM_INSECURE_TLS}).addExtended()

	cmd := exec.CommandContext(getCtx, r.dir, r.args...)
	cmd.Env = append(cmd.Env, r.opt.RepoEnvs.Slice()...)
	logger.Infof("[Cmd] %s", cmd.String())
	output, _ := cmd.CombinedOutput()
	logger.Debugf("[restic] prune result: %s", string(output))

	return string(output), nil
}

func (r *Restic) Stats() (*StatsContainer, error) {
	var getCtx, cancel = context.WithCancel(r.ctx)
	defer cancel()

	r.addCommand([]string{"stats", "--mode", "raw-data", PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS}).addExtended().addRequestTimeout()

	opts := utils.CommandOptions{
		Path: r.dir,
		Args: r.args,
		Envs: r.opt.RepoEnvs.Kv(),
	}

	c := utils.NewCommand(getCtx, opts)

	var stats *StatsContainer
	var errorMsg RESTIC_ERROR_MESSAGE

	go func() {
		for {
			select {
			case res, ok := <-c.Ch:
				if !ok {
					return
				}
				if res == nil || len(res) == 0 {
					continue
				}

				var msg = string(res)
				logger.Debugf("[restic] stats %s message: %s", r.opt.RepoName, msg)
				if err := json.Unmarshal(res, &stats); err != nil {
					errorMsg = RESTIC_ERROR_MESSAGE(string(msg))
					c.Cancel()
					return
				}
			case <-r.ctx.Done():
				return
			}
		}
	}()

	_, err := c.Run()
	if err != nil {
		return nil, err
	}
	if errorMsg != "" {
		return nil, fmt.Errorf(errorMsg.Error())
	}

	if stats == nil {
		return nil, fmt.Errorf("stats %s not found", r.opt.RepoName)
	}

	return stats, nil
}

func (r *Restic) Tag(snapshotId string, tags []string) error {
	r.addCommand([]string{"tag"}).resetTags(tags).addSnapshotId(snapshotId)

	cmd := exec.CommandContext(context.Background(), r.dir, r.args...)
	cmd.Env = append(cmd.Env, r.opt.RepoEnvs.Slice()...)

	logger.Infof("[Cmd] %s", cmd.String())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func (r *Restic) Backup(folder string, files []string, filePathPrefix string, tags []string, traceId string, dryRun bool, progressChan chan float64) (*SummaryOutput, error) {
	var filesPath, err = r.formatBackupFiles(files)
	if err != nil {
		return nil, fmt.Errorf("invalid backup file list path, error: %v", err.Error())
	}

	var cmds = []string{"backup"}

	if r.opt.BackupType == constants.BackupTypeApp {
		if filesPath == nil {
			return nil, fmt.Errorf("backup app but files is empty")
		}
		for _, file := range filesPath {
			cmds = append(cmds, "--files-from", file)
		}
	} else {
		if folder != "" {
			cmds = append(cmds, folder)
		}
	}
	if dryRun {
		cmds = append(cmds, "-n")
	}

	cmds = append(cmds, r.opt.SetLimitUploadRate(), PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS)

	r.addCommand(cmds).
		addExtended().
		addRequestTimeout().
		addTags(tags)

	opts := utils.CommandOptions{
		Path: r.dir,
		Args: r.args,
		Envs: r.opt.RepoEnvs.Kv(),
	}

	c := utils.NewCommand(r.ctx, opts)

	var prevPercent float64
	var finished bool
	var summary *SummaryOutput
	var errorMsg RESTIC_ERROR_MESSAGE
	var continued bool

	go func() {
		for {
			select {
			case <-r.ctx.Done():
				logger.Infof("[restic] backup canceled, traceId: %s", traceId)
				errorMsg = ERROR_MESSAGE_BACKUP_CANCELED
				return
			case res, ok := <-c.Ch:
				if !ok {
					return
				}
				if res == nil || len(res) == 0 {
					continue
				}
				status := messagePool.Get()
				if err := json.Unmarshal(res, status); err != nil {
					var msg = string(res)
					logger.Errorf("[restic] backup %s error message: %s, traceId: %s", r.opt.RepoName, msg, traceId)
					messagePool.Put(status)

					errorMsg, continued = r.formatErrorMessage(msg)
					if continued {
						continue
					}
					c.Cancel()
					return
				}
				switch status.MessageType {
				case "status":
					switch {
					case math.Abs(status.PercentDone-0.0) < tolerance:
						logger.Infof(PRINT_START_MESSAGE, status.TotalFiles, utils.FormatBytes(status.TotalBytes))
						progressChan <- status.PercentDone
					case math.Abs(status.PercentDone-1.0) < tolerance:
						if !finished {
							logger.Infof(PRINT_FINISH_MESSAGE, status.TotalFiles, utils.FormatBytes(status.TotalBytes))
							finished = true
							progressChan <- status.PercentDone
						}
					default:
						if prevPercent != 0 && prevPercent != status.PercentDone {
							logger.Infof(PRINT_PROGRESS_MESSAGE,
								status.GetPercentDone(),
								status.FilesDone,
								status.TotalFiles,
								utils.FormatBytes(status.BytesDone),
								utils.FormatBytes(status.TotalBytes),
								r.fileNameTidy(status.CurrentFiles, filePathPrefix))
							progressChan <- status.PercentDone
						}
						prevPercent = status.PercentDone
					}
				case "error":
					var continued bool
					errObj := new(ErrorUpdate)
					if err := json.Unmarshal(res, &errObj); err != nil {
						errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
					} else {
						errorMsg, continued = r.formatErrorMessage(errObj.Error.Message)
						if continued {
							continue
						}
					}
					messagePool.Put(status)
					c.Cancel()
					return
				case "summary":
					if err := json.Unmarshal(res, &summary); err != nil {
						logger.Errorf("[restic] backup %s error summary unmarshal message: %s, traceId: %s", r.opt.RepoName, string(res), traceId)
						messagePool.Put(status)
						errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
						c.Cancel()
						return
					}
					messagePool.Put(status)
					return
				}
				messagePool.Put(status)
			}
		}
	}()

	_, err = c.Run()
	if err != nil {
		return nil, err
	}
	if errorMsg != "" {
		return nil, fmt.Errorf(errorMsg.Error())
	}
	return summary, nil
}

func (r *Restic) Repair() error {
	backoff := wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    10,
	}

	var e error

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		res, err := r.repairIndex()
		if err != nil {
			logger.Errorf("[restic] repair %s error: %s", r.opt.RepoName, err)
			return err
		}

		if strings.Contains(res, ERROR_MESSAGE_LOCKED.Error()) {
			r.Unlock()
			return fmt.Errorf("retry")
		}

		if strings.Contains(res, ERROR_MESSAGE_WRONG_PASSWORD_OR_NO_KEY_FOUND.Error()) {
			e = errors.New(ERROR_MESSAGE_WRONG_PASSWORD.Error())
		}
		return nil
	}); err != nil {
		return err
	}
	if e != nil {
		return e
	}
	return nil
}

func (r *Restic) repairIndex() (string, error) {
	r.addCommand([]string{"repair", "index", PARAM_INSECURE_TLS}).addExtended()

	opts := utils.CommandOptions{
		Path:  r.dir,
		Args:  r.args,
		Envs:  r.opt.RepoEnvs.Kv(),
		Print: true,
	}

	c := utils.NewCommand(r.ctx, opts)

	sb := new(strings.Builder)
	go func() {
		for {
			select {
			case res, ok := <-c.Ch:
				if !ok {
					return
				}
				if res == nil || len(res) == 0 {
					continue
				}
				logger.Debugf("[restic] repair %s message: %s", r.opt.RepoName, string(res))
				sb.WriteString(string(res) + "\n")
			case <-r.ctx.Done():
				return
			}
		}
	}()

	_, err := c.Run()
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (r *Restic) Unlock() (string, error) {
	var getCtx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	r.addCommand([]string{"unlock", "--remove-all", PARAM_INSECURE_TLS}).addExtended()

	opts := utils.CommandOptions{
		Path: r.dir,
		Args: r.args,
		Envs: r.opt.RepoEnvs.Kv(),
	}
	c := utils.NewCommand(getCtx, opts)
	sb := new(strings.Builder)

	go func() {
		for {
			select {
			case res, ok := <-c.Ch:
				if !ok {
					return
				}
				if res == nil || len(res) == 0 {
					continue
				}
				logger.Debugf("[restic] unlock %s message: %s", r.opt.RepoName, string(res))
				sb.WriteString(string(res) + "\n")
			case <-r.ctx.Done():
				return
			}
		}
	}()

	_, err := c.Run()
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (r *Restic) GetSnapshot(snapshotId string) (*Snapshot, error) {
	var getCtx, cancel = context.WithCancel(r.ctx)
	defer cancel()

	r.addCommand([]string{"snapshots", PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS, snapshotId}).addExtended().addRequestTimeout()

	opts := utils.CommandOptions{
		Path: r.dir,
		Args: r.args,
		Envs: r.opt.RepoEnvs.Kv(),
	}

	c := utils.NewCommand(getCtx, opts)

	var summary []*Snapshot
	var errorMsg RESTIC_ERROR_MESSAGE

	go func() {
		for {
			select {
			case res, ok := <-c.Ch:
				if !ok {
					return
				}
				if res == nil || len(res) == 0 {
					continue
				}

				var msg = string(res)
				logger.Debugf("[restic] snapshots %s message: %s", r.opt.RepoName, msg)
				if err := json.Unmarshal(res, &summary); err != nil {
					errorMsg = RESTIC_ERROR_MESSAGE(string(msg))
					c.Cancel()
					return
				}
			case <-r.ctx.Done():
				return
			}
		}
	}()

	_, err := c.Run()
	if err != nil {
		return nil, err
	}
	if errorMsg != "" {
		return nil, fmt.Errorf(errorMsg.Error())
	}

	if summary == nil || len(summary) == 0 {
		return nil, fmt.Errorf("snapshot %s not found", snapshotId)
	}

	return summary[0], nil
}

type SnapshotList []*Snapshot

func (l SnapshotList) First() *Snapshot {
	return l[0]
}

func (l SnapshotList) Len() int {
	return len(l)
}

func (l SnapshotList) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Time", "Host", "Tags", "Path", "Size"})

	for _, s := range l {
		var data = []string{s.ShortId, s.Time, s.Hostname, strings.Join(s.Tags, "\n"), strings.Join(s.Paths, "\n"), utils.FormatBytes(uint64(s.Summary.TotalBytesProcessed))}
		table.Append(data)
	}
	table.Render()
}

func (r *Restic) GetSnapshots(tags []string) (*SnapshotList, error) {
	var restoreCtx, cancel = context.WithCancel(r.ctx)
	defer cancel()

	r.addCommand([]string{"snapshots", PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS}).addTags(tags).addExtended().addRequestTimeout()

	opts := utils.CommandOptions{
		Path: r.dir,
		Args: r.args,
		Envs: r.opt.RepoEnvs.Kv(),
	}

	c := utils.NewCommand(restoreCtx, opts)

	var summary *SnapshotList
	var errorMsg RESTIC_ERROR_MESSAGE

	go func() {
		for {
			select {
			case res, ok := <-c.Ch:
				if !ok {
					return
				}
				if res == nil || len(res) == 0 {
					continue
				}

				var msg = string(res)
				logger.Debugf("[restic] snapshots %s message: %s", r.opt.RepoName, msg)
				if strings.Contains(msg, "Fatal: ") {
					errorMsg, _ = r.formatErrorMessage(msg)
					return
				}
				if err := json.Unmarshal(res, &summary); err != nil {
					errorMsg = RESTIC_ERROR_MESSAGE(r.trimError(err.Error()))
					return
				}
			case <-r.ctx.Done():
				return
			}
		}
	}()

	_, err := c.Run()
	if err != nil {
		return nil, err
	}
	if errorMsg != "" {
		return nil, fmt.Errorf(errorMsg.Error())
	}
	if summary == nil || len(*summary) == 0 {
		return nil, fmt.Errorf("snapshots not found")
	}
	return summary, nil
}

func (r *Restic) Restore(phase int, total int, snapshotId string, subfolder string, target string, progressChan chan float64) (*RestoreSummaryOutput, error) {
	if subfolder != "" {
		subfolder = fmt.Sprintf("%s:%s", snapshotId, subfolder)
	} else {
		subfolder = snapshotId
	}
	r.addCommand([]string{"restore", r.opt.SetLimitDownloadRate(), "-t", target, "-v=3", PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS, subfolder}).addExtended().addRequestTimeout()

	// var restoreCtx, cancel = context.WithCancel(r.ctx)
	// defer cancel()
	opts := utils.CommandOptions{
		Path: r.dir,
		Args: r.args,
		Envs: r.opt.RepoEnvs.Kv(),
	}

	c := utils.NewCommand(r.ctx, opts)

	var prevPercent float64
	var started bool
	var finished bool
	var summary *RestoreSummaryOutput
	var errorMsg RESTIC_ERROR_MESSAGE
	var continued bool

	go func() {
		for {
			select {
			case <-r.ctx.Done():
				logger.Infof("[restic] restore canceled")
				errorMsg = ERROR_MESSAGE_RESTORE_CANCELED
				return
			case res, ok := <-c.Ch:
				if !ok {
					return
				}
				if res == nil || len(res) == 0 {
					continue
				}

				status := restoreMessagePool.Get()
				if err := json.Unmarshal(res, status); err != nil {
					var msg = string(res)
					logger.Debugf("[restic] restore %s error message: %s", r.opt.RepoName, msg)
					restoreMessagePool.Put(status)

					errorMsg, continued = r.formatErrorMessage(msg)
					if continued {
						continue
					}

					c.Cancel()
					return
				}
				switch status.MessageType {
				case "status":
					switch {
					case math.Abs(status.PercentDone-0.0) < tolerance:
						if !started {
							logger.Infof(PRINT_RESTORE_START_MESSAGE, status.TotalFiles, utils.FormatBytes(status.TotalBytes))
							started = true
							progressChan <- status.GetPercentDone(phase, total)
						}
					case math.Abs(status.PercentDone-1.0) < tolerance:
						if !finished {
							logger.Infof(PRINT_RESTORE_FINISH_MESSAGE, snapshotId, status.TotalFiles, status.FilesRestored, utils.FormatBytes(status.TotalBytes), utils.FormatBytes(status.BytesRestored))
							finished = true
							progressChan <- status.GetPercentDone(phase, total)
						}
					default:
						if prevPercent != status.PercentDone {
							logger.Infof(PRINT_RESTORE_PROGRESS_MESSAGE,
								fmt.Sprintf("%.2f%%", status.GetPercentDone(phase, total)*100),
								status.FilesRestored,
								status.TotalFiles,
								utils.FormatBytes(status.BytesRestored),
								utils.FormatBytes(status.TotalBytes),
							)
							progressChan <- status.GetPercentDone(phase, total)
						}
						prevPercent = status.PercentDone
					}
				case "verbose_status":
					rvu := new(RestoreVerboseUpdate)
					if err := json.Unmarshal(res, &rvu); err != nil {
						errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
						restoreMessagePool.Put(status)
						c.Cancel()
						return
					}
					logger.Infof(PRINT_RESTORE_ITEM, rvu.Item, utils.FormatBytes(rvu.Size))
				case "error":
					errObj := new(ErrorUpdate)
					if err := json.Unmarshal(res, &errObj); err != nil {
						errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
					} else {
						errorMsg = RESTIC_ERROR_MESSAGE(errObj.Error.Message)
					}
					restoreMessagePool.Put(status)
					c.Cancel()
					return
				case "summary":
					if err := json.Unmarshal(res, &summary); err != nil {
						logger.Debugf("[restic] restore %s error summary unmarshal message: %s", r.opt.RepoName, string(res))
						restoreMessagePool.Put(status)
						errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
						c.Cancel()
						return
					}
					restoreMessagePool.Put(status)
					return
				}
				restoreMessagePool.Put(status)

			}
		}
	}()

	_, err := c.Run()
	if err != nil {
		return nil, err
	}
	if errorMsg != "" {
		return nil, fmt.Errorf(errorMsg.Error())
	}

	return summary, nil
}

func (r *Restic) fileNameTidy(f []string, prefix string) []string {
	if f == nil || len(f) == 0 {
		return f
	}

	var res []string
	for _, file := range f {
		res = append(res, strings.TrimPrefix(file, prefix))
	}

	return res
}

func (r *Restic) addCommand(args []string) *Restic {
	r.args = args
	return r
}

func (r *Restic) addSnapshotId(snapshotId string) *Restic {
	r.args = append(r.args, snapshotId)
	return r
}

func (r *Restic) resetTags(tags []string) *Restic {
	if tags == nil {
		return r
	}
	for _, tag := range tags {
		if tag != "" {
			r.args = append(r.args, "--add", tag)
		}
	}
	return r
}

func (r *Restic) addTags(tags []string) *Restic {
	if tags == nil {
		return r
	}
	for _, tag := range tags {
		if tag != "" {
			r.args = append(r.args, "--tag", tag)
		}
	}
	return r
}

func (r *Restic) addExtended() *Restic {
	var cloudName = strings.ToLower(r.opt.CloudName)
	if cloudName == constants.CloudTencentName {
		r.args = append(r.args, "-o", "s3.bucket-lookup=dns", "-o", fmt.Sprintf("s3.region=%s", r.opt.RegionId))
	}
	return r
}

func (r *Restic) addRequestTimeout() *Restic {
	r.args = append(r.args, "--stuck-request-timeout", "60s")
	return r
}

func (r *Restic) formatBackupFiles(files []string) ([]string, error) {
	if files == nil || len(files) == 0 {
		return nil, nil
	}

	var res []string
	for _, f := range files {
		pt, err := filepath.Abs(f)
		if err != nil {
			return nil, err
		}
		res = append(res, pt)
	}

	return res, nil
}

func (r *Restic) trimError(s string) string {
	return strings.ReplaceAll(s, "Fatal: ", "")
}

func (r *Restic) formatErrorMessage(msg string) (RESTIC_ERROR_MESSAGE, bool) {
	var errorMsg RESTIC_ERROR_MESSAGE
	var continued bool = false
	switch {
	case strings.Contains(msg, ERROR_MESSAGE_ALREADY_INITIALIZED.Error()), strings.Contains(msg, ERROR_MESSAGE_CONFIG_FILE_ALREADY_EXISTS.Error()):
		errorMsg = MESSAGE_REPOSITORY_ALREADY_INITIALIZED // Init
	case strings.Contains(msg, ERROR_MESSAGE_SNAPSHOT_NOT_FOUND.Error()):
		errorMsg = ERROR_MESSAGE_SNAPSHOT_NOT_FOUND
	case strings.Contains(msg, ERROR_MESSAGE_TOKEN_EXPIRED.Error()),
		strings.Contains(msg, ERROR_MESSAGE_COS_TOKEN_EXPIRED.Error()):
		errorMsg = RESTIC_ERROR_MESSAGE(ERROR_MESSAGE_TOKEN_EXPIRED.ToLower())
	case strings.Contains(msg, ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE.Error()):
		errorMsg = ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_SERVER_MISBEHAVING.Error()):
		errorMsg = ERROR_MESSAGE_SERVER_MISBEHAVING_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_ACCESS_DENIED.Error()):
		errorMsg = ERROR_MESSAGE_ACCESS_DENIED_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_NO_SUCH_DEVICE.Error()):
		errorMsg = ERROR_MESSAGE_NO_SUCH_DEVICE_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_HOST_IS_DOWN.Error()):
		errorMsg = ERROR_MESSAGE_HOST_IS_DOWN_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_COS_ACCOUNT_ARREARS.Error()):
		errorMsg = ERROR_MESSAGE_COS_ACCOUNT_ARREARS_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_REPOSITORY_BE_DAMAGED.Error()):
		errorMsg = ERROR_MESSAGE_REPOSITORY_BE_DAMAGED_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_RESOURCE_TEMPORARILY_UNAVAILABLE.Error()):
		errorMsg = ERROR_MESSAGE_HOST_IS_DOWN_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE.Error()):
		errorMsg = ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_NO_SPACE_LEFT_ON_DEVICE.Error()):
		errorMsg = ERROR_MESSAGE_NO_SPACE_LEFT_ON_DEVICE_MESSAGE
	case strings.Contains(msg, ERROR_MESSAGE_WRONG_PASSWORD_OR_NO_KEY_FOUND.Error()):
		errorMsg = ERROR_MESSAGE_WRONG_PASSWORD
	case strings.Contains(msg, ERROR_MESSAGE_REPOSITORY_DOES_NOT_EXIST.Error()):
		errorMsg = ERROR_MESSAGE_REPOSITORY_DOES_NOT_EXIST_MESSAGE
	case strings.Contains(msg, "path") && strings.Contains(msg, "not found"),
		strings.Contains(msg, ERROR_MESSAGE_FILES_NOT_FOUND.Error()),
		strings.Contains(msg, ERROR_MESSAGE_NO_SUCH_FILE_OR_DIRECTORY.Error()):
		continued = true
	default:
		errorMsg = RESTIC_ERROR_MESSAGE(msg)
	}

	return errorMsg, continued
}

var messagePool *statusMessagePool

type statusMessagePool struct {
	pool sync.Pool
}

var restoreMessagePool *restoreStatusMessagePool

type restoreStatusMessagePool struct {
	pool sync.Pool
}

func init() {
	messagePool = NewResticMessagePool()
	restoreMessagePool = NewResticRestoreMessagePool()
}

func NewResticRestoreMessagePool() *restoreStatusMessagePool {
	return &restoreStatusMessagePool{
		pool: sync.Pool{
			New: func() any {
				obj := new(RestoreStatusUpdate)
				return obj
			},
		},
	}
}

func (r *restoreStatusMessagePool) Get() *RestoreStatusUpdate {
	if obj := r.pool.Get(); obj != nil {
		return obj.(*RestoreStatusUpdate)
	}
	var obj = new(RestoreStatusUpdate)
	r.Put(obj)
	return obj
}

func (r *restoreStatusMessagePool) Put(obj *RestoreStatusUpdate) {
	r.pool.Put(obj)
}

func NewResticMessagePool() *statusMessagePool {
	return &statusMessagePool{
		pool: sync.Pool{
			New: func() any {
				obj := new(StatusUpdate)
				return obj
			},
		},
	}
}

func (r *statusMessagePool) Get() *StatusUpdate {
	if obj := r.pool.Get(); obj != nil {
		return obj.(*StatusUpdate)
	}
	var obj = new(StatusUpdate)
	r.Put(obj)
	return obj
}

func (r *statusMessagePool) Put(obj *StatusUpdate) {
	r.pool.Put(obj)
}
