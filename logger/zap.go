package logger

import (
	"invtools/utils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger      *zap.Logger
	LoggerSugar *zap.SugaredLogger
)

func InitLog() {
	var cfg zap.Config
	var runMode = "debug"
	if runMode == "release" {
		cfg = zap.NewProductionConfig()
		cfg.DisableCaller = true
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.LevelKey = "level"
		cfg.EncoderConfig.NameKey = "name"
		cfg.EncoderConfig.MessageKey = "msg"
		cfg.EncoderConfig.CallerKey = "caller"
		cfg.EncoderConfig.StacktraceKey = "stacktrace"
	}

	cfg.Encoding = "json"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{"logs/invtools.log"}
	//cfg.ErrorOutputPaths = []string{"logs/error.log"}
	utils.Mkdir("logs")
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	Logger = l
	LoggerSugar = Logger.Sugar()
}

func Errorf(template string, args ...interface{}) {
	LoggerSugar.Errorf(template, args)
}
func ErrorfWithEnv(debug bool, template string, args ...interface{}) {
	if debug {
		LoggerSugar.Errorf(template, args)
	}
}

func Debugf(template string, args ...interface{}) {
	LoggerSugar.Debugf(template, args)
}

func DebugfWithEnv(debug bool, template string, args ...interface{}) {
	if debug {
		LoggerSugar.Debugf(template, args)
	}
}

func Infof(template string, args ...interface{}) {
	LoggerSugar.Infof(template, args)
}

func InfofWithEnv(debug bool, template string, args ...interface{}) {
	if debug {
		LoggerSugar.Infof(template, args)
	}
}
