package logrotate

import (
	"fmt"
	"gopkg.in/robfig/cron.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	logFileDir    = "/logs"
	logFilePrefix = "operator"
)

var logFilePath = path.Join(logFileDir, fmt.Sprintf("%s.log", logFilePrefix))
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
func StartRotate() error {
	rotator := cron.New()
	if _, err := rotator.AddFunc("0 0 1 * * ?", rotateLog); err != nil {
		return err
	}
	rotator.Start()
	return nil
}

func rotateLog() {
	in, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	filePath := path.Join(logFileDir, fmt.Sprintf("%s.%s.log", logFilePrefix, time.Now().AddDate(0, 0, -1).Format("2006-01-02")))
	if err := ioutil.WriteFile(filePath, in, 0644); err != nil {
		fmt.Println(err)
		return
	}

	logger.Info(fmt.Sprintf("Log backup succeeded (%s)", filePath))
	if err := os.Truncate(logFilePath, 0); err != nil {
		fmt.Println(err)
		return
	}

	if logFile != nil {
		if _, err := logFile.Seek(0, io.SeekStart); err != nil {
			fmt.Println(err)
			return
		}
	}
}
