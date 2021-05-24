// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package logger

import (
	"log"

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
	Fatal(msg string, keyValues ...interface{})
}

type zapLogger struct {
	logger *zap.SugaredLogger
}

// make sure to implement Logger interface
var _ Logger = (*zapLogger)(nil)

func (zl *zapLogger) Debug(msg string, keyValues ...interface{}) { zl.logger.Debugw(msg, keyValues...) }
func (zl *zapLogger) Info(msg string, keyValues ...interface{})  { zl.logger.Infow(msg, keyValues...) }
func (zl *zapLogger) Warn(msg string, keyValues ...interface{})  { zl.logger.Warnw(msg, keyValues...) }
func (zl *zapLogger) Error(msg string, keyValues ...interface{}) { zl.logger.Errorw(msg, keyValues...) }
func (zl *zapLogger) Panic(msg string, keyValues ...interface{}) { zl.logger.Panicw(msg, keyValues...) }
func (zl *zapLogger) Fatal(msg string, keyValues ...interface{}) { zl.logger.Fatalw(msg, keyValues...) }

// Config for Logger
type Config struct {
	Debug bool
	Level zapcore.Level
}

// New create production logger
func New() Logger {
	return NewWithConfig(Config{
		Debug: true,
	})
}

// NewWithConfig creates a new logger with config
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

// NewNop creates no-op logger
func NewNop() Logger {
	return &zapLogger{zap.NewNop().Sugar()}
}

var myLogger Logger

// Set sets a global logger
func Set(logger Logger) {
	myLogger = logger
}

func Debug(msg string, keyValues ...interface{}) { myLogger.Debug(msg, keyValues...) }
func Info(msg string, keyValues ...interface{})  { myLogger.Info(msg, keyValues...) }
func Warn(msg string, keyValues ...interface{})  { myLogger.Warn(msg, keyValues...) }
func Error(msg string, keyValues ...interface{}) { myLogger.Error(msg, keyValues...) }
func Panic(msg string, keyValues ...interface{}) { myLogger.Panic(msg, keyValues...) }
func Fatal(msg string, keyValues ...interface{}) { myLogger.Fatal(msg, keyValues...) }

func init() {
	Set(NewNop())
}
