package logger

import (
	"errors"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
	Empty    = ""
)

const (
	consoleEncoding = "console"
	jsonEncoding    = "json"
)

var ErrNoLog = errors.New("logger can not be empty")

type Logger struct {
	*zap.Logger
}

func New(env string) (Logger, error) {
	var output []string
	var outputErr []string
	var encoding string
	var logLvl zapcore.Level
	switch env {
	case EnvLocal:
		logLvl = zap.DebugLevel
		encoding = consoleEncoding
		output = []string{"stdout"}
		outputErr = []string{"stderr"}
	case EnvDev:
		logLvl = zap.DebugLevel
		encoding = jsonEncoding
		output = []string{"stdout", "./logs/log.log"}
		outputErr = []string{"stderr", "./logs/err.log"}
	case EnvProd:
		logLvl = zap.InfoLevel
		encoding = jsonEncoding
		output = []string{"stdout", "./logs/log.log"}
		outputErr = []string{"stderr", "./logs/err.log"}
	default:
		return Logger{}, ErrNoLog
	}

	config := zap.Config{
		Level:    zap.NewAtomicLevelAt(logLvl),
		Encoding: encoding,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.TimeEncoderOfLayout("2006 Jan 02 15:04:05 -0700 MST"),

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,

			LineEnding: "\n",
		},
		OutputPaths:      output,
		ErrorOutputPaths: outputErr,
	}

	core, err := getCore(logLvl, config)

	logg := zap.New(core)
	if err != nil {
		return Logger{}, err
	}

	return Logger{logg}, nil
}

func getCore(logLvl zapcore.Level, config zap.Config) (zapcore.Core, error) {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zap.ErrorLevel && lvl >= logLvl
	})

	ws, err := toMultiSyncer(config.OutputPaths)
	if err != nil {
		return nil, err
	}

	wsErr, err := toMultiSyncer(config.ErrorOutputPaths)
	if err != nil {
		return nil, err
	}
	var core zapcore.Core

	switch config.Encoding {
	case jsonEncoding:
		core = zapcore.NewTee(
			zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), ws, levelEnabler),
			zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), wsErr, highPriority),
		)
	case consoleEncoding:
		core = zapcore.NewTee(
			zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), ws, levelEnabler),
			zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), wsErr, highPriority),
		)
	}

	return core, nil
}

func toMultiSyncer(files []string) (zapcore.WriteSyncer, error) {
	w := make([]zapcore.WriteSyncer, 0, len(files))

	for _, f := range files {
		switch f {
		case "stderr":
			w = append(w, zapcore.AddSync(os.Stderr))
		case "stdout":
			w = append(w, zapcore.AddSync(os.Stdout))
		default:
			if err := os.MkdirAll(filepath.Dir(f), os.ModePerm); err != nil {
				return nil, err
			}

			file, err := os.OpenFile(f, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o666)
			if err != nil {
				return nil, err
			}
			w = append(w, zapcore.AddSync(file))
		}
	}

	ws := zapcore.NewMultiWriteSyncer(w...)
	return ws, nil
}
