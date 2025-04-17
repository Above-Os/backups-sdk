package restic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

type RESTIC_ERROR_MESSAGE string

const (
	SUCCESS_MESSAGE_REPAIR_INDEX                 RESTIC_ERROR_MESSAGE = "adding pack file to index"
	ERROR_MESSAGE_UNABLE_TO_OPEN_REPOSITORY      RESTIC_ERROR_MESSAGE = "unable to open repository at"
	ERROR_MESSAGE_TOKEN_EXPIRED                  RESTIC_ERROR_MESSAGE = "The provided token has expired"
	ERROR_MESSAGE_COS_TOKEN_EXPIRED              RESTIC_ERROR_MESSAGE = "The Access Key Id you provided does not exist in our records"
	ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE     RESTIC_ERROR_MESSAGE = "unable to open config file: Stat: 400 Bad Request"
	ERROR_MESSAGE_CONFIG_INVALID                 RESTIC_ERROR_MESSAGE = "config invalid, please chek repository or authorization config"
	ERROR_MESSAGE_LOCKED                         RESTIC_ERROR_MESSAGE = "repository is already locked by"
	ERROR_MESSAGE_ALREADY_INITIALIZED            RESTIC_ERROR_MESSAGE = "repository master key and config already initialized"
	ERROR_MESSAGE_SNAPSHOT_NOT_FOUND             RESTIC_ERROR_MESSAGE = "no matching ID found for prefix"
	ERROR_MESSAGE_CONFIG_FILE_ALREADY_EXISTS     RESTIC_ERROR_MESSAGE = "config file already exists"
	ERROR_MESSAGE_WRONG_PASSWORD_OR_NO_KEY_FOUND RESTIC_ERROR_MESSAGE = "wrong password or no key found"
	ERROR_MESSAGE_REPOSITORY_DOES_NOT_EXIST      RESTIC_ERROR_MESSAGE = "repository does not exist: unable to open config file"
	ERROR_MESSAGE_BACKUP_CANCELED                RESTIC_ERROR_MESSAGE = "backup canceled"
	ERROR_MESSAGE_RESTORE_CANCELED               RESTIC_ERROR_MESSAGE = "restore canceled"
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
	RepoName          string
	RepoSuffix        string
	CloudName         string
	RegionId          string
	SnapshotId        string
	Path              string
	LimitDownloadRate string
	LimitUploadRate   string

	RepoEnvs *ResticEnvs
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
		return nil, err
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

	var outerr string
	output, err := cmd.CombinedOutput()
	if err != nil {
		var errmsg = string(output)
		switch {
		case strings.Contains(errmsg, ERROR_MESSAGE_ALREADY_INITIALIZED.Error()), strings.Contains(errmsg, ERROR_MESSAGE_CONFIG_FILE_ALREADY_EXISTS.Error()):
			outerr = MESSAGE_REPOSITORY_ALREADY_INITIALIZED
		// case strings.Contains(errmsg, ERROR_MESSAGE_UNABLE_TO_OPEN_REPOSITORY.Error()):
		// 	outerr = MESSAGE_TOKEN_EXPIRED
		default:
			outerr = errmsg
		}
	}

	if outerr != "" {
		return "", errors.New(outerr)
	}

	return string(output), nil
}

func (r *Restic) Stats() (*StatsContainer, error) {
	var getCtx, cancel = context.WithCancel(r.ctx)
	defer cancel()

	r.addCommand([]string{"stats", PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS}).addExtended().addRequestTimeout()

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

func (r *Restic) Backup(folder string, filePathPrefix string, tags []string, progressCallback func(percentDone float64)) (*SummaryOutput, error) {
	r.addCommand([]string{"backup", folder, r.opt.SetLimitUploadRate(), PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS}).
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

	go func() {
		for {
			select {
			case <-r.ctx.Done():
				logger.Infof("[restic] backup canceled")
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
					logger.Debugf("[restic] backup %s error message: %s", r.opt.RepoName, msg)
					messagePool.Put(status)
					switch {
					case strings.Contains(msg, ERROR_MESSAGE_TOKEN_EXPIRED.Error()),
						strings.Contains(msg, ERROR_MESSAGE_COS_TOKEN_EXPIRED.Error()):
						errorMsg = ERROR_MESSAGE_TOKEN_EXPIRED
						c.Cancel()
						return
					case strings.Contains(msg, ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE.Error()):
						errorMsg = ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE
						c.Cancel()
						return
					default:
						errorMsg = RESTIC_ERROR_MESSAGE(msg)
						c.Cancel()
						return
					}
				}
				switch status.MessageType {
				case "status":
					switch {
					case math.Abs(status.PercentDone-0.0) < tolerance:
						logger.Infof(PRINT_START_MESSAGE, status.TotalFiles, utils.FormatBytes(status.TotalBytes))
						progressCallback(status.PercentDone)
					case math.Abs(status.PercentDone-1.0) < tolerance:
						if !finished {
							logger.Infof(PRINT_FINISH_MESSAGE, status.TotalFiles, utils.FormatBytes(status.TotalBytes))
							finished = true
							progressCallback(status.PercentDone)
						}
					default:
						if prevPercent != status.PercentDone {
							logger.Infof(PRINT_PROGRESS_MESSAGE,
								status.GetPercentDone(),
								status.FilesDone,
								status.TotalFiles,
								utils.FormatBytes(status.BytesDone),
								utils.FormatBytes(status.TotalBytes),
								r.fileNameTidy(status.CurrentFiles, filePathPrefix))
							progressCallback(status.PercentDone)
						}
						prevPercent = status.PercentDone
					}
				case "summary":
					if err := json.Unmarshal(res, &summary); err != nil {
						logger.Debugf("[restic] backup %s error summary unmarshal message: %s", r.opt.RepoName, string(res))
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

	_, err := c.Run()
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

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		res, err := r.repairIndex()
		if err != nil {
			return err
		}

		if strings.Contains(res, ERROR_MESSAGE_LOCKED.Error()) {
			r.Unlock()
			return fmt.Errorf("retry")
		}
		return nil
	}); err != nil {
		return err
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
	r.addCommand([]string{"unlock", "--remove-all", PARAM_INSECURE_TLS}).addExtended()

	opts := utils.CommandOptions{
		Path: r.dir,
		Args: r.args,
		Envs: r.opt.RepoEnvs.Kv(),
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
					switch {
					case strings.Contains(msg, ERROR_MESSAGE_SNAPSHOT_NOT_FOUND.Error()):
						errorMsg = ERROR_MESSAGE_SNAPSHOT_NOT_FOUND
						return
					case strings.Contains(msg, ERROR_MESSAGE_WRONG_PASSWORD_OR_NO_KEY_FOUND.Error()):
						errorMsg = ERROR_MESSAGE_WRONG_PASSWORD_OR_NO_KEY_FOUND
						return
					case strings.Contains(msg, ERROR_MESSAGE_REPOSITORY_DOES_NOT_EXIST.Error()):
						errorMsg = ERROR_MESSAGE_REPOSITORY_DOES_NOT_EXIST
						return
					default:
						errorMsg = RESTIC_ERROR_MESSAGE(msg)
						return
					}
				}
				if err := json.Unmarshal(res, &summary); err != nil {
					errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
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

func (r *Restic) Restore(snapshotId string, uploadPath string, target string, progressCallback func(percentDone float64)) (*RestoreSummaryOutput, error) {
	r.addCommand([]string{"restore", r.opt.SetLimitDownloadRate(), "-t", target, "-v=3", PARAM_JSON_OUTPUT, PARAM_INSECURE_TLS, fmt.Sprintf("%s:%s", snapshotId, uploadPath)}).addExtended().addRequestTimeout()

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

					switch {
					case strings.Contains(msg, ERROR_MESSAGE_TOKEN_EXPIRED.Error()),
						strings.Contains(msg, ERROR_MESSAGE_COS_TOKEN_EXPIRED.Error()):
						errorMsg = ERROR_MESSAGE_TOKEN_EXPIRED
						c.Cancel()
						return
					case strings.Contains(msg, ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE.Error()):
						errorMsg = ERROR_MESSAGE_UNABLE_TO_OPEN_CONFIG_FILE
						c.Cancel()
						return
					default:
						errorMsg = RESTIC_ERROR_MESSAGE(msg)
						c.Cancel()
						return
					}
				}
				switch status.MessageType {
				case "status":
					switch {
					case math.Abs(status.PercentDone-0.0) < tolerance:
						if !started {
							logger.Infof(PRINT_RESTORE_START_MESSAGE, status.TotalFiles, utils.FormatBytes(status.TotalBytes))
							started = true
							progressCallback(status.PercentDone)
						}
					case math.Abs(status.PercentDone-1.0) < tolerance:
						if !finished {
							logger.Infof(PRINT_RESTORE_FINISH_MESSAGE, snapshotId, status.TotalFiles, status.FilesRestored, utils.FormatBytes(status.TotalBytes), utils.FormatBytes(status.BytesRestored))
							finished = true
							progressCallback(status.PercentDone)
						}
					default:
						if prevPercent != status.PercentDone {
							logger.Infof(PRINT_RESTORE_PROGRESS_MESSAGE,
								status.GetPercentDone(),
								status.FilesRestored,
								status.TotalFiles,
								utils.FormatBytes(status.BytesRestored),
								utils.FormatBytes(status.TotalBytes),
							)
							progressCallback(status.PercentDone)
						}
						prevPercent = status.PercentDone
					}
				case "verbose_status":
					rvu := new(RestoreVerboseUpdate)
					if err := json.Unmarshal(res, &rvu); err != nil {
						errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
						c.Cancel()
						return
					}
					logger.Infof(PRINT_RESTORE_ITEM, rvu.Item, utils.FormatBytes(rvu.Size))
				case "summary":
					if err := json.Unmarshal(res, &summary); err != nil {
						logger.Debugf("[restic] restore %s error summary unmarshal message: %s", r.opt.RepoName, string(res))
						restoreMessagePool.Put(status)
						errorMsg = RESTIC_ERROR_MESSAGE(err.Error())
						c.Cancel()
						return
					}
					restoreMessagePool.Put(status)
					progressCallback(1.0)
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
	var cloudName = r.opt.CloudName
	if cloudName == constants.CloudTencentName {
		r.args = append(r.args, "-o", "s3.bucket-lookup=dns", "-o", fmt.Sprintf("s3.region=%s", r.opt.RegionId))
	}
	return r
}

func (r *Restic) addRequestTimeout() *Restic {
	r.args = append(r.args, "--stuck-request-timeout", "120s")
	return r
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
