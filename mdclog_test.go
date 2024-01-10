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

package golog

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getTestLogger returns a logger instance where
// the output is directed to a byte buffer instead
// of stdout
func getTestLogger(t *testing.T) (*MdcLogger, *bytes.Buffer) {
	logbuffer := new(bytes.Buffer)
	logger, err := initLogger("foo", logbuffer)
	assert.Nil(t, err)
	return logger, logbuffer
}

func TestLogInitDoesNotReturnAnError(t *testing.T) {
	_, err := InitLogger("foo")
	assert.Nil(t, err, "create failed")
}

func TestDebugFunctionLogsCorrectString(t *testing.T) {
	logger, logbuffer := getTestLogger(t)
	logger.Debug("test debug")
	logstr := logbuffer.String()
	assert.Contains(t, logstr, "crit\":\"DEBUG\",\"id\":\"foo\",\"mdc\":{},\"msg\":\"test debug\"}\n")
}

func TestInfoFunctionLogsCorrectString(t *testing.T) {
	logger, logbuffer := getTestLogger(t)
	logger.Info("test info")
	logstr := logbuffer.String()
	assert.Contains(t, logstr, "crit\":\"INFO\",\"id\":\"foo\",\"mdc\":{},\"msg\":\"test info\"}\n")
}

func TestWarningLogsCorrectString(t *testing.T) {
	logger, logbuffer := getTestLogger(t)
	logger.Warning("test warn")
	logstr := logbuffer.String()
	assert.Contains(t, logstr, "crit\":\"WARNING\",\"id\":\"foo\",\"mdc\":{},\"msg\":\"test warn\"}\n")
}

func TestErrorFunctionLogsCorrectString(t *testing.T) {
	logger, logbuffer := getTestLogger(t)
	logger.Error("test err")
	logstr := logbuffer.String()
	assert.Contains(t, logstr, "crit\":\"ERROR\",\"id\":\"foo\",\"mdc\":{},\"msg\":\"test err\"}\n")
}

func TestLogFunctionLogsCorrectString(t *testing.T) {
	logger, logbuffer := getTestLogger(t)
	logger.Log(ERR, "test err")
	logstr := logbuffer.String()
	assert.Contains(t, logstr, "crit\":\"ERROR\",\"id\":\"foo\",\"mdc\":{},\"msg\":\"test err\"}\n")
}

func TestFormatWithMdcReturnsJsonFormatedString(t *testing.T) {
	logger, _ := InitLogger("foo")
	logger.MdcAdd("foo", "bar")
	logstr, err := logger.formatLog(INFO, "test2")
	assert.Nil(t, err, "formatLog fails")
	v := make(map[string]interface{})
	err = json.Unmarshal(logstr, &v)
	assert.Equal(t, "INFO", v["crit"])
	assert.Equal(t, "test2", v["msg"])
	assert.Equal(t, "foo", v["id"])
	expectedmdc := map[string]interface{}{"foo": "bar"}
	assert.Equal(t, expectedmdc, v["mdc"])
}

func TestMdcAddIsOk(t *testing.T) {
	logger, _ := InitLogger("foo")
	logger.MdcAdd("foo", "bar")
	val, ok := logger.MdcGet("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
}

func TestMdcRemoveWorks(t *testing.T) {
	logger, _ := InitLogger("foo")
	logger.MdcAdd("foo", "bar")
	val, ok := logger.MdcGet("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	logger.MdcRemove("foo")
	val, ok = logger.MdcGet("foo")
	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestRemoveNonExistentMdcDoesNotCrash(t *testing.T) {
	logger, _ := InitLogger("foo")
	logger.MdcRemove("foo")
}

func TestMdcCleanRemovesAllMdcs(t *testing.T) {
	logger, _ := InitLogger("foo")
	logger.MdcAdd("foo1", "bar")
	logger.MdcAdd("foo2", "bar")
	logger.MdcAdd("foo3", "bar")
	logger.MdcClean()
	_, ok := logger.MdcGet("foo1")
	assert.False(t, ok)
	_, ok = logger.MdcGet("foo2")
	assert.False(t, ok)
	_, ok = logger.MdcGet("foo3")
	assert.False(t, ok)
}

func TestLevelStringsGetterWorks(t *testing.T) {
	assert.Equal(t, "ERROR", levelString(ERR))
	assert.Equal(t, "WARNING", levelString(WARN))
	assert.Equal(t, "INFO", levelString(INFO))
	assert.Equal(t, "DEBUG", levelString(DEBUG))
}

func TestDefaultLoggingLevelIsDebug(t *testing.T) {
	logger, _ := InitLogger("foo")
	assert.Equal(t, DEBUG, logger.LevelGet())
}

func TestLevelGetReturnsWhatWasSet(t *testing.T) {
	logger, _ := InitLogger("foo")
	logger.LevelSet(ERR)
	assert.Equal(t, ERR, logger.LevelGet())
}

func TestDebugLogIsNotWrittenIfCurrentLevelIsInfo(t *testing.T) {
	logger, logbuffer := getTestLogger(t)
	logger.LevelSet(INFO)
	logger.Debug("fooo")
	assert.Empty(t, logbuffer.String())
}

func TestLogFormatWithMdcArray(t *testing.T) {
	logger, _ := InitLogger("app")
	logFileMonitor := 0
	logger.Mdclog_format_initialize(logFileMonitor)
	logstr, err := logger.formatLog(INFO, "test")
	assert.Nil(t, err, "formatLog fails")
	v := make(map[string]interface{})
	err = json.Unmarshal(logstr, &v)
	assert.Equal(t, "INFO", v["crit"])
	assert.Equal(t, "test", v["msg"])
	assert.Equal(t, "app", v["id"])
	_, ok := logger.MdcGet("SYSTEM_NAME")
	assert.True(t, ok)
	_, ok = logger.MdcGet("HOST_NAME")
	assert.True(t, ok)
	_, ok = logger.MdcGet("SERVICE_NAME")
	assert.True(t, ok)
	_, ok = logger.MdcGet("CONTAINER_NAME")
	assert.True(t, ok)
	_, ok = logger.MdcGet("POD_NAME")
	assert.True(t, ok)
}

func TestLogLevelConfigFileParse(t *testing.T) {
	logger, _ := InitLogger("app")
	d1 := []byte("log-level:WARN\n\n")
	err := ioutil.WriteFile("/tmp/log-file", d1, 0644)
	assert.Nil(t, err, "Failed to create tmp log-file")
	os.Setenv("CONFIG_MAP_NAME", "/tmp/log-file")
	logFileMonitor := 1
	logger.Mdclog_format_initialize(logFileMonitor)
	assert.Equal(t, WARN, logger.LevelGet())
	_, ok := logger.MdcGet("PID")
	assert.True(t, ok)
	logger.Mdclog_format_initialize(logFileMonitor)
}
