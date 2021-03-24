// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package mlog

import (
	"encoding/json"
	"fmt"
	"os"
)

// defaultLog manually encodes the log to STDERR, providing a basic, default logging implementation
// before mlog is fully configured.
func defaultLog(level, msg string, fields ...Field) {
	log := struct {
		Level   string  `json:"level"`
		Message string  `json:"msg"`
		Fields  []Field `json:"fields,omitempty"`
	}{
		level,
		msg,
		fields,
	}

	if b, err := json.Marshal(log); err != nil {
		fmt.Fprintf(os.Stderr, `{"level":"error","msg":"failed to encode log message"}%s`, "\n")
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", b)
	}
}

func defaultLogS(level, msg string, keyValuePairs ...interface{}) {
	//log := struct { `json:"msg"`
	//	Fields  []Field `json:"fields,omitempty"`
	//}{
	//	level,
	//	msg,
	//	fields,
	//}

	log := map[string]interface{}{
		"level": level,
		"msg":   msg,
	}

	var key string
	i := 0
	for _, arg := range keyValuePairs {
		if key == "" {
			if kStr, ok := arg.(string); !ok {
				log[fmt.Sprintf("params%d", i)] = arg
				i++
				continue
			} else {
				key = kStr
			}
		} else {
			log[key] = arg
			key = ""
		}
	}
	if key != "" {
		log[fmt.Sprintf("params%d", i)] = key
	}

	if b, err := json.Marshal(log); err != nil {
		fmt.Fprintf(os.Stderr, `{"level":"error","msg":"failed to encode log message"}%s`, "\n")
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", b)
	}
}

func defaultDebugLog(msg string, fields ...Field) {
	defaultLog("debug", msg, fields...)
}

func defaultInfoLog(msg string, fields ...Field) {
	defaultLog("info", msg, fields...)
}

func defaultWarnLog(msg string, fields ...Field) {
	defaultLog("warn", msg, fields...)
}

func defaultErrorLog(msg string, fields ...Field) {
	defaultLog("error", msg, fields...)
}

func defaultCriticalLog(msg string, fields ...Field) {
	// We map critical to error in zap, so be consistent.
	defaultLog("error", msg, fields...)
}

func defaultDebugSLog(msg string, keyValuePairs ...interface{}) {
	defaultLogS("debug", msg, keyValuePairs...)
}

func defaultInfoSLog(msg string, keyValuePairs ...interface{}) {
	defaultLogS("info", msg, keyValuePairs...)
}

func defaultWarnSLog(msg string, keyValuePairs ...interface{}) {
	defaultLogS("warn", msg, keyValuePairs...)
}

func defaultErrorSLog(msg string, keyValuePairs ...interface{}) {
	defaultLogS("error", msg, keyValuePairs...)
}

func defaultCriticalSLog(msg string, keyValuePairs ...interface{}) {
	// We map critical to error in zap, so be consistent.
	defaultLogS("error", msg, keyValuePairs...)
}
