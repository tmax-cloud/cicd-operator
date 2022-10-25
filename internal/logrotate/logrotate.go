/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package logrotate

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"gopkg.in/robfig/cron.v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	defaultLogFileDir = "/logs"
	logFilePrefix     = "operator"
)

var logDir = defaultLogFileDir
var logFilePath = path.Join(logDir, fmt.Sprintf("%s.log", logFilePrefix))
var logger = ctrl.Log.WithName("logrotate")
var logFile *os.File

// LogFile opens a file for the log
func LogFile() (*os.File, error) {
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return nil, err
	}
	logFile = file
	return file, nil
}

// StartRotate starts a cronjob to rotate the log
func StartRotate(spec string) error {
	rotator := cron.New()
	if _, err := rotator.AddFunc(spec, rotateLog); err != nil {
		return err
	}
	rotator.Start()
	return nil
}

func rotateLog() {
	in, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		logger.Error(err, "read log error")
		return
	}

	filePath := path.Join(logDir, fmt.Sprintf("%s.%s.log", logFilePrefix, time.Now().AddDate(0, 0, -1).Format("2006-01-02")))
	if err := ioutil.WriteFile(filePath, in, 0644); err != nil {
		logger.Error(err, "write log error")
		return
	}

	logger.Info(fmt.Sprintf("Log backup succeeded (%s)", filePath))
	if err := os.Truncate(logFilePath, 0); err != nil {
		logger.Error(err, "truncate log error")
		return
	}

	if logFile != nil {
		if _, err := logFile.Seek(0, io.SeekStart); err != nil {
			logger.Error(err, "seek log error")
			return
		}
	}
}
