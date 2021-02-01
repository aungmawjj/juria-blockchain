// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package logger

import (
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger supports structured logging
type Logger interface {
	Debug(msg string, keyValues ...interface{})
	Info(msg string, keyValues ...interface{})
	Warn(msg string, keyValues ...interface{})
	Error(msg string, keyValues ...interface{})
	Panic(msg string, keyValues ...interface{})
}

type zapLogger struct {
	logger *zap.SugaredLogger
}

var _ Logger = (*zapLogger)(nil)

func (zl *zapLogger) Debug(msg string, keyValues ...interface{}) { zl.logger.Debugw(msg, keyValues) }
func (zl *zapLogger) Info(msg string, keyValues ...interface{})  { zl.logger.Infow(msg, keyValues) }
func (zl *zapLogger) Warn(msg string, keyValues ...interface{})  { zl.logger.Warnw(msg, keyValues) }
func (zl *zapLogger) Error(msg string, keyValues ...interface{}) { zl.logger.Errorw(msg, keyValues) }
func (zl *zapLogger) Panic(msg string, keyValues ...interface{}) { zl.logger.Panicw(msg, keyValues) }

// Config for Logger
type Config struct {
	Debug bool
	Level zapcore.Level
}

// New create production logger
func New() Logger {
	return NewWithConfig(Config{false, 0})
}

// NewWithConfig returns a new logger
func NewWithConfig(cfg Config) Logger {
	var (
		logger *zap.Logger
		err    error
	)
	if cfg.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction(zap.IncreaseLevel(cfg.Level))
	}
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	return &zapLogger{logger.Sugar()}
}

var logger Logger
var once sync.Once

// Init creates a global logger
func Init(l Logger) {
	once.Do(func() {
		logger = l
	})
}

// Instance returns global Logger
func Instance() Logger {
	if logger == nil {
		log.Fatalf("logger isn't initialized")
	}
	return logger
}
