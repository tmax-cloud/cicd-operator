package logrotate

import (
	"fmt"
	"gopkg.in/robfig/cron.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	LogFileDir    = "/logs"
	LogFilePrefix = "operator"
)

var LogFilePath = path.Join(LogFileDir, fmt.Sprintf("%s.log", LogFilePrefix))

var logger = ctrl.Log.WithName("logrotate")

var logFile *os.File

func LogFile() (*os.File, error) {
	file, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return nil, err
	}
	logFile = file
	return file, nil
}

func StartRotate() error {
	rotator := cron.New()
	if _, err := rotator.AddFunc("0 0 1 * * ?", rotateLog); err != nil {
		return err
	}
	rotator.Start()
	return nil
}

func rotateLog() {
	in, err := ioutil.ReadFile(LogFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	filePath := path.Join(LogFileDir, fmt.Sprintf("%s.%s.log", LogFilePrefix, time.Now().AddDate(0, 0, -1).Format("2006-01-02")))
	if err := ioutil.WriteFile(filePath, in, 0644); err != nil {
		fmt.Println(err)
		return
	}

	logger.Info(fmt.Sprintf("Log backup succeeded (%s)", filePath))
	if err := os.Truncate(LogFilePath, 0); err != nil {
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
