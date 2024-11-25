package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Zando74/GopherFS/metadata-service/config"
	"github.com/rs/zerolog"
)

type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

type Logger struct {
	once     sync.Once
	instance *Logger
	logger   *zerolog.Logger
}

var _ Interface = (*Logger)(nil)

func NewLogger(level string) *Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		l = zerolog.ErrorLevel
	case "warn":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
}

func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg("debug", message, args...)
}

func (l *Logger) Debugf(message string, args ...interface{}) {
	l.logger.Debug().Msgf(message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log(message, args...)
}

func (l *Logger) Infof(message string, args ...interface{}) {
	l.logger.Info().Msgf(message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(message, args...)
}

func (l *Logger) Warningf(message string, args ...interface{}) {
	l.logger.Warn().Msgf(message, args...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	if l.logger.GetLevel() == zerolog.DebugLevel {
		l.Debug(message, args...)
	}

	l.msg("error", message, args...)
}

func (l *Logger) Errorf(message string, args ...interface{}) {
	l.logger.Error().Msgf(message, args...)
}

func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg("fatal", message, args...)

	os.Exit(1)
}

func (l *Logger) log(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Info().Msg(message)
	} else {
		l.logger.Info().Msgf(message, args...)
	}
}

func (l *Logger) msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(msg.Error(), args...)
	case string:
		l.log(msg, args...)
	default:
		l.log(fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}

func (cfg *Logger) GetInstance() *Logger {
	cfg.once.Do(func() {
		cfg.instance = NewLogger(config.ConfigSingleton.GetInstance().Log.Level)
	})

	return cfg.instance
}

var LoggerSingleton Logger
