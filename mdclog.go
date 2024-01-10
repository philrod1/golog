/*
 *  Copyright (c) 2019 AT&T Intellectual Property.
 *  Copyright (c) 2018-2019 Nokia.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 * This source code is part of the near-RT RIC (RAN Intelligent Controller)
 * platform project (RICP).
 */

// Package golog implements a simple structured logging with MDC (Mapped Diagnostics Context) support.
package golog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/utils/inotify"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Level is a type define for the logging level.
type Level int

const (
	// ERR is an error level log entry.
	ERR Level = 1
	// WARN is a warning level log entry.
	WARN Level = 2
	// INFO is an info level log entry.
	INFO Level = 3
	// DEBUG is a debug level log entry.
	DEBUG Level = 4
)

// MdcLogger is the logger instance, created with InitLogger() function.
type MdcLogger struct {
	proc      string
	writer    io.Writer
	mdc       map[string]string
	mutex     sync.Mutex
	level     Level
	init_done int
}

type logEntry struct {
	Ts   int64             `json:"ts"`
	Crit string            `json:"crit"`
	Id   string            `json:"id"`
	Mdc  map[string]string `json:"mdc"`
	Msg  string            `json:"msg"`
}

func levelString(level Level) string {
	switch level {
	case ERR:
		return "ERROR"
	case WARN:
		return "WARNING"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	default:
		return ""
	}
}

func getTime() int64 {
	ns := time.Time.UnixNano(time.Now())
	return ns / int64(time.Millisecond)
}

func (l *MdcLogger) formatLog(level Level, msg string) ([]byte, error) {
	log := logEntry{getTime(), levelString(level), l.proc, l.mdc, msg}
	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(log)
	return buf.Bytes(), err
}

func initLogger(proc string, writer io.Writer) (*MdcLogger, error) {
	return &MdcLogger{proc: proc, writer: writer, mdc: make(map[string]string), level: DEBUG, init_done: 0}, nil
}

// InitLogger is the init routine which returns a new logger instance.
// The program identity is given as a parameter. The identity
// is added to every log writing.
// The function returns a new instance or an error.
func InitLogger(proc string) (*MdcLogger, error) {
	return initLogger(proc, os.Stdout)
}

// Log is the basic logging function to write a log message with
// the given level
func (l *MdcLogger) Log(level Level, formatMsg string, a ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.level < level {
		return
	}
	log, err := l.formatLog(level, fmt.Sprintf(formatMsg, a...))
	if err == nil {
		l.writer.Write(log)
	}
}

// Error is the "error" level logging function.
func (l *MdcLogger) Error(formatMsg string, a ...interface{}) {
	l.Log(ERR, formatMsg, a...)
}

// Warning is the "warning" level logging function.
func (l *MdcLogger) Warning(formatMsg string, a ...interface{}) {
	l.Log(WARN, formatMsg, a...)
}

// Info is the "info" level logging function.
func (l *MdcLogger) Info(formatMsg string, a ...interface{}) {
	l.Log(INFO, formatMsg, a...)
}

// Debug is the "debug" level logging function.
func (l *MdcLogger) Debug(formatMsg string, a ...interface{}) {
	l.Log(DEBUG, formatMsg, a...)
}

// LevelSet sets the current logging level.
// Log writings with less significant level are discarded.
func (l *MdcLogger) LevelSet(level Level) {
	l.level = level
}

// LevelGet returns the current logging level.
func (l *MdcLogger) LevelGet() Level {
	return l.level
}

// MdcAdd adds a new MDC key value pair to the logger.
func (l *MdcLogger) MdcAdd(key string, value string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.mdc[key] = value
}

// MdcRemove removes an MDC key from the logger.
func (l *MdcLogger) MdcRemove(key string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.mdc, key)
}

// MdcGet gets the value of an MDC from the logger.
// The function returns the value string and a boolean
// which tells if the key was found or not.
func (l *MdcLogger) MdcGet(key string) (string, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	val, ok := l.mdc[key]
	return val, ok
}

// MdcClean removes all MDC keys from the logger.
func (l *MdcLogger) MdcClean() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.mdc = make(map[string]string)
}

func (l *MdcLogger) MdcUpdate(key string, value string) {
	_, ok := l.MdcGet(key)
	if ok {
		l.MdcRemove(key)
	}
	l.MdcAdd(key, value)
}

func (l *MdcLogger) ParseFileContent(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	for _, lineData := range strings.Split(string(data), "\n") {
		if strings.Contains(lineData, "log-level:") {
			var level = ERR
			strList := strings.Split(lineData, ":")
			if strings.Contains(strings.ToUpper(strList[1]), "DEBUG") {
				level = DEBUG
			} else if strings.Contains(strings.ToUpper(strList[1]), "INFO") {
				level = INFO
			} else if strings.Contains(strings.ToUpper(strList[1]), "ERR") {
				level = ERR
			} else if strings.Contains(strings.ToUpper(strList[1]), "WARN") {
				level = WARN
			}
			l.LevelSet(level)
		}
	}
}

func (l *MdcLogger) watch_changes(watcher *inotify.Watcher, fileName string) {
	for {
		select {
		case ev := <-watcher.Event:
			if strings.Contains(ev.Name, filepath.Dir(fileName)) {
				l.ParseFileContent(fileName)
			}
		case err := <-watcher.Error:
			fmt.Println("error:", err)
		}
	}
}

func (l *MdcLogger) readEnvVar(envKey string) string {
	envValue, provided := os.LookupEnv(envKey)
	if !provided {
		envValue = ""
	}
	return envValue
}

func (l *MdcLogger) Mdclog_format_initialize(logFileMonitor int) int {
	ret := -1
	logFields := []string{"SYSTEM_NAME", "HOST_NAME", "SERVICE_NAME", "CONTAINER_NAME", "POD_NAME"}
	for _, envKey := range logFields {
		envValue := l.readEnvVar(envKey)
		l.MdcUpdate(envKey, envValue)
	}
	l.MdcUpdate("PID", strconv.Itoa(os.Getpid()))
	if logFileMonitor > 0 {
		watchPath := l.readEnvVar("CONFIG_MAP_NAME")
		_, err := os.Stat(watchPath)
		if !os.IsNotExist(err) {
			if l.init_done == 0 {
				l.mutex.Lock()
				l.init_done = 1
				l.mutex.Unlock()
				watcher, err := inotify.NewWatcher()
				if err != nil {
					return -1
				}
				err = watcher.AddWatch(filepath.Dir(watchPath), syscall.IN_CLOSE_WRITE|syscall.IN_CREATE|syscall.IN_CLOSE)
				if err != nil {
					return -1
				}
				l.ParseFileContent(watchPath)
				go l.watch_changes(watcher, watchPath)
				ret = 0
			}
		}
	}
	return ret
}
