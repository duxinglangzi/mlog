// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package mlog

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const (
	// Very verbose messages for debugging specific issues
	LevelDebug = "debug"
	// Default log level, informational
	LevelInfo = "info"
	// Warnings are messages about possible issues
	LevelWarn = "warn"
	// Errors are messages about things we know are problems
	LevelError = "error"
)

// Type and function aliases from zap to limit the libraries scope into MM code
type Field = zapcore.Field

var Int64 = zap.Int64
var Int32 = zap.Int32
var Int = zap.Int
var Uint32 = zap.Uint32
var String = zap.String
var Any = zap.Any
var Err = zap.Error
var Bool = zap.Bool
var Duration = zap.Duration

type LoggerConfiguration struct {
	EnableConsole   bool
	ConsoleFile     string
	FileLevelConfig map[string]string
	
	InfoLevelFiles  *[]string
	WarnLevelFiles  *[]string
	ErrorLevelFiles *[]string
	DebugLevelFiles *[]string
	
	// 日志保留天数 默认7天
	MaxHistoryDays *uint
	
	// ConsoleLevel    string
	// ConsoleJson     bool
	// EnableFile      bool
	// FileJson        bool
	// FileLevel       string
	// FileLocation    string
}

type Logger struct {
	zap          *zap.Logger
	consoleLevel zap.AtomicLevel
	fileLevel    zap.AtomicLevel
}

func getZapLevel(level string) zapcore.Level {
	switch level {
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func makeEncoder(json bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if json {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func NewLogger(config *LoggerConfiguration) *Logger {
	cores := []zapcore.Core{}
	logger := &Logger{
		// consoleLevel: zap.NewAtomicLevelAt(getZapLevel(config.ConsoleLevel)),
		// fileLevel:    zap.NewAtomicLevelAt(getZapLevel(config.FileLevel)),
	}
	// 增加日志保留天数控制
	var options []rotatelogs.Option
	if config.MaxHistoryDays != nil {
		options = append(options, rotatelogs.WithMaxAge(time.Duration(*config.MaxHistoryDays)*24*time.Hour))
	} else {
		options = append(options, rotatelogs.WithMaxAge(7*24*time.Hour))
	}
	
	// 控制台输出
	cores = append(cores, zapcore.NewCore(makeEncoder(true), zapcore.Lock(os.Stderr), zapcore.DebugLevel))
	
	// 单独设置 console 的日志文件
	if config.EnableConsole && len(config.ConsoleFile) > 0 {
		logs, _ := rotatelogs.New(strings.Replace(config.ConsoleFile, ".log", "", -1)+"-%Y%m%d.log", options...)
		level := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.DebugLevel
		})
		core := zapcore.NewCore(makeEncoder(true), zapcore.AddSync(logs), level)
		cores = append(cores, core)
	}
	
	// 循环 map 设置日志文件
	for key, val := range config.FileLevelConfig {
		logs, _ := rotatelogs.New(strings.Replace(val, ".log", "", -1)+"-%Y%m%d.log", options...)
		writer := zapcore.AddSync(logs)
		zapLevel := getZapLevel(strings.ToLower(key))
		level := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == zapLevel
		})
		core := zapcore.NewCore(makeEncoder(true), writer, level)
		cores = append(cores, core)
	}
	
	// if config.EnableFile {
	// 	writer := zapcore.AddSync(&lumberjack.Logger{
	// 		Filename: config.FileLocation,
	// 		MaxSize:  100,
	// 		Compress: true,
	// 	})
	// 	core := zapcore.NewCore(makeEncoder(config.FileJson), writer, logger.fileLevel)
	// 	cores = append(cores, core)
	// }
	
	combinedCore := zapcore.NewTee(cores...)
	logger.zap = zap.New(combinedCore,
		zap.AddCaller(),
	)
	return logger
}

// func (l *Logger) ChangeLevels(config *LoggerConfiguration) {
// 	l.consoleLevel.SetLevel(getZapLevel(config.ConsoleLevel))
// 	l.fileLevel.SetLevel(getZapLevel(config.FileLevel))
// }

func (l *Logger) SetConsoleLevel(level string) {
	l.consoleLevel.SetLevel(getZapLevel(level))
}

func (l *Logger) With(fields ...Field) *Logger {
	newlogger := *l
	newlogger.zap = newlogger.zap.With(fields...)
	return &newlogger
}

func (l *Logger) StdLog(fields ...Field) *log.Logger {
	return zap.NewStdLog(l.With(fields...).zap.WithOptions(getStdLogOption()))
}

// StdLogAt returns *log.Logger which writes to supplied zap logger at required level.
func (l *Logger) StdLogAt(level string, fields ...Field) (*log.Logger, error) {
	return zap.NewStdLogAt(l.With(fields...).zap.WithOptions(getStdLogOption()), getZapLevel(level))
}

// StdLogWriter returns a writer that can be hooked up to the output of a golang standard logger
// anything written will be interpreted as log entries accordingly
func (l *Logger) StdLogWriter() io.Writer {
	newLogger := *l
	newLogger.zap = newLogger.zap.WithOptions(zap.AddCallerSkip(4), getStdLogOption())
	f := newLogger.Info
	return &loggerWriter{f}
}

func (l *Logger) WithCallerSkip(skip int) *Logger {
	newlogger := *l
	newlogger.zap = newlogger.zap.WithOptions(zap.AddCallerSkip(skip))
	return &newlogger
}

func (l *Logger) WithNewZapCoreLogger(core zapcore.Core) *Logger {
	newlogger := *l
	newlogger.zap = zap.New(zapcore.NewTee(l.zap.Core(), core),
		zap.AddCaller(),
	)
	return &newlogger
}

// Made for the plugin interface, wraps mlog in a simpler interface
// at the cost of performance
func (l *Logger) Sugar() *SugarLogger {
	return &SugarLogger{
		wrappedLogger: l,
		zapSugar:      l.zap.Sugar(),
	}
}

func (l *Logger) Debug(message string, fields ...Field) {
	l.zap.Debug(message, fields...)
}

func (l *Logger) Info(message string, fields ...Field) {
	l.zap.Info(message, fields...)
}

func (l *Logger) Warn(message string, fields ...Field) {
	l.zap.Warn(message, fields...)
}

func (l *Logger) Error(message string, fields ...Field) {
	l.zap.Error(message, fields...)
}

func (l *Logger) Critical(message string, fields ...Field) {
	l.zap.Error(message, fields...)
}
