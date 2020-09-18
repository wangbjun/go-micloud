package zlog

import (
	"fmt"
	"go-micloud/configs"
	"go-micloud/pkg/color"
	"go-micloud/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

var Logger *zap.Logger

func init() {
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})
	logFile := configs.Conf.Section("APP").Key("LOG_FILE").String()
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    500, // megabytes
		MaxBackups: 0,
		MaxAge:     28, // days
		LocalTime:  true,
	})
	sync := zapcore.AddSync(writer)

	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewTee(zapcore.NewCore(jsonEncoder, sync, infoLevel))

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	defer logger.Sync()

	Logger = logger
}

func PrintInfo(msg string) {
	fmt.Printf(color.Green(time.Now().Format("2006-01-02 15:04:05")+" #%s\n"), msg)
	Logger.Sugar().With("time", time.Now().Format(utils.YmdHis)).Info(msg)
}

func PrintError(msg string) {
	fmt.Printf(color.Red(time.Now().Format("2006-01-02 15:04:05")+" #%s\n"), msg)
	Logger.Sugar().With("time", time.Now().Format(utils.YmdHis)).Error(msg)
}

func Info(msg string) {
	Logger.Sugar().With("time", time.Now().Format(utils.YmdHis)).Info(msg)
}

func Warn(msg string) {
	Logger.Sugar().With("time", time.Now().Format(utils.YmdHis)).Warn(msg)
}

func Error(msg string) {
	Logger.Sugar().With("time", time.Now().Format(utils.YmdHis)).Error(msg)
}
