package logger

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"olares.com/backups-sdk/pkg/constants"
)

var logger *zap.SugaredLogger

var FatalMessagePrefix = "[FATAL] "

func InitLogger(consoleLogTruncate bool) {
	var user, err = user.Current()
	if err != nil {
		panic(errors.New("get current user failed"))
	}

	var homeDir, jsonLogDir string

	if err = godotenv.Load(constants.OlaresReleaseFile); err != nil {
		homeDir = user.HomeDir
		jsonLogDir = path.Join(homeDir, constants.DefaultBaseDir, constants.DefaultLogsDir)
	} else {
		envOlaresBaseDir := os.Getenv(constants.ENV_OLARES_BASE_DIR)
		if envOlaresBaseDir == "" {
			homeDir = path.Join(user.HomeDir, constants.DefaultBaseDir)
		} else {
			homeDir = envOlaresBaseDir
		}
		jsonLogDir = path.Join(homeDir, constants.DefaultLogsDir)
	}

	found, err := isDirExist(jsonLogDir)
	if err != nil {
		fmt.Println("log dir found error", err)
	}
	if !found {
		err := os.MkdirAll(jsonLogDir, 0755)
		if err != nil {
			fmt.Println("create log dir error", err)
			os.Exit(1)
		}
	}

	jsonLogFileName := path.Join(jsonLogDir, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02_15-04-05")))
	jsonLogFile, err := os.OpenFile(jsonLogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}

	consolePriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.DebugLevel
	})
	jsonLogFilePriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
	})

	fileEncoder := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleDebugging := zapcore.Lock(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEncoderConfig), consoleDebugging, consolePriority),
		zapcore.NewCore(zapcore.NewJSONEncoder(fileEncoder), zapcore.AddSync(jsonLogFile), jsonLogFilePriority),
	)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.FatalLevel)).Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return logger
}

func SetLogger(l *zap.SugaredLogger) {
	logger = l
}

func isDirExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func getLevel(level string) (l zapcore.Level) {
	switch level {
	case "debug":
		l = zap.DebugLevel
	case "info":
		l = zap.InfoLevel
	case "warn":
		l = zap.WarnLevel
	case "error":
		l = zap.ErrorLevel
	case "panic":
		l = zap.PanicLevel
	case "fatal":
		l = zap.FatalLevel
	default:
		l = zap.InfoLevel
	}
	return
}

func Debug(args ...any) {
	logger.Debug(args...)
}

func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Debugw(msg string, args ...any) {
	logger.Debugw(msg, args...)
}

func Info(args ...any) {
	logger.Info(args...)
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Infow(msg string, args ...any) {
	logger.Infow(msg, args...)
}

func InfoInstallationProgress(format string, args ...any) {
	logger.Infof("[INFO] "+format, args...)
}

func Warn(args ...any) {
	logger.Warn(args...)
}

func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Warnw(msg string, args ...any) {
	logger.Warnw(msg, args...)
}

func Error(args ...any) {
	logger.Error(args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

func Errorw(msg string, args ...any) {
	logger.Errorw(msg, args...)
}

func DPanic(args ...any) {
	logger.DPanic(args...)
}

func DPanicf(format string, args ...any) {
	logger.DPanicf(format, args...)
}

func DPanicw(msg string, args ...any) {
	logger.DPanicw(msg, args...)
}

func Panic(args ...any) {
	logger.Panic(args...)
}

func Panicf(format string, args ...any) {
	logger.Panicf(format, args...)
}

func Panicw(msg string, args ...any) {
	logger.Panicw(msg, args...)
}

func Fatal(args ...any) {
	var newArgs []any
	newArgs = append(newArgs, FatalMessagePrefix)
	newArgs = append(newArgs, args...)
	logger.Fatal(newArgs...)
}

func Fatalf(format string, args ...any) {
	logger.Fatalf(FatalMessagePrefix+format, args...)
}

func Fatalw(msg string, args ...any) {
	logger.Fatalw(msg, args...)
}
