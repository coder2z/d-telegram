package xlog

import (
	"github.com/coder2z/d-telegram/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
	"time"
)

var (
	logger *zap.Logger
	one    sync.Once
)

func getLogWriter(cfg *config.Config) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.Log.WriteSyncer.Filename,
		MaxSize:    cfg.Log.WriteSyncer.MaxSize,
		MaxAge:     cfg.Log.WriteSyncer.MaxAge,
		MaxBackups: cfg.Log.WriteSyncer.MaxBackups,
		LocalTime:  cfg.Log.WriteSyncer.LocalTime,
		Compress:   cfg.Log.WriteSyncer.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func defaultZapConfig(cfg *config.Config) *zapcore.EncoderConfig {
	var encodeTime zapcore.TimeEncoder = timeEncoder
	if cfg.Log.Debug {
		encodeTime = timeDebugEncoder
	}
	return &zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "lv",
		TimeKey:        "ts",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     encodeTime,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.Unix())
}
func timeDebugEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func GetLogger() *zap.Logger {
	one.Do(func() {
		cfg := config.Get()
		zapOptions := make([]zap.Option, 0)
		zapOptions = append(zapOptions, zap.AddStacktrace(zap.DPanicLevel))
		if cfg.Log.AddCaller {
			zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(cfg.Log.CallerSkip))
		}
		writeSyncer := getLogWriter(cfg)
		if cfg.Log.Debug {
			writeSyncer = zap.CombineWriteSyncers(os.Stdout, writeSyncer)
		}
		lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
		if err := lv.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
			panic(err)
		}
		core := zapcore.NewCore(
			func() zapcore.Encoder {
				if cfg.Log.Debug {
					return zapcore.NewConsoleEncoder(*defaultZapConfig(cfg))
				}
				return zapcore.NewJSONEncoder(*defaultZapConfig(cfg))
			}(),
			writeSyncer,
			lv,
		)

		logger = zap.New(core, zapOptions...)
	})
	return logger
}
