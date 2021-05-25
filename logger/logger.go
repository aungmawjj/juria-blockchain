// Copyright (C) 2021 Aung Maw
// Licensed under the GNU General Public License v3.0

package logger

import (
	"go.uber.org/zap"
)

var myLogger *zap.SugaredLogger

// Set sets a global logger
func Set(logger *zap.SugaredLogger) {
	myLogger = logger
}

func I() *zap.SugaredLogger {
	return myLogger
}

func init() {
	Set(zap.NewNop().Sugar())
}
