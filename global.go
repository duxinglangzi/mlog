// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package mlog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *Logger

func InitGlobalLogger(logger *Logger) {
	glob := *logger
	glob.zap = glob.zap.WithOptions(zap.AddCallerSkip(1))
	globalLogger = &glob
	Debug = globalLogger.Debug
	Info = globalLogger.Info
	Warn = globalLogger.Warn
	Error = globalLogger.Error
	Critical = globalLogger.Critical
	sugerLogger := globalLogger.Sugar()
	DebugW = sugerLogger.Debug
	InfoW = sugerLogger.Info
	WarnW = sugerLogger.Warn
	ErrorW = sugerLogger.Error
	CriticalW = sugerLogger.Critical

}

func GlobalLoggerWithSkipper(skip int) *Logger {
	newLogger := *globalLogger
	newLogger.zap = newLogger.zap.WithOptions(zap.AddCallerSkip(1))
	return &newLogger
}

func RedirectStdLog(logger *Logger) {
	zap.RedirectStdLogAt(logger.zap.With(zap.String("source", "stdlog")).WithOptions(zap.AddCallerSkip(-2)), zapcore.ErrorLevel)
}

type LogFunc func(string, ...Field)
type LogSFunc func(string, ...interface{})

// DON'T USE THIS Modify the level on the app logger
func GloballyDisableDebugLogForTest() {
	globalLogger.consoleLevel.SetLevel(zapcore.ErrorLevel)
}

// DON'T USE THIS Modify the level on the app logger
func GloballyEnableDebugLogForTest() {
	globalLogger.consoleLevel.SetLevel(zapcore.DebugLevel)
}

//var LogFileOnly = globalLogger.LogFileOnly
var Debug LogFunc = defaultDebugLog
var Info LogFunc = defaultInfoLog
var Warn LogFunc = defaultWarnLog
var Error LogFunc = defaultErrorLog
var Critical LogFunc = defaultCriticalLog

var DebugW LogSFunc = defaultDebugSLog
var InfoW LogSFunc = defaultInfoSLog
var WarnW LogSFunc = defaultWarnSLog
var ErrorW LogSFunc = defaultErrorSLog
var CriticalW LogSFunc = defaultCriticalSLog

var Format = fmt.Sprintf
