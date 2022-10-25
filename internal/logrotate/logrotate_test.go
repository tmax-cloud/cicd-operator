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
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/test"
)

func TestLogFile(t *testing.T) {
	tc := map[string]struct {
		dirRO   bool
		fileErr bool

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {},
		"mkdirErr": {
			dirRO:        true,
			errorOccurs:  true,
			errorMessage: "operator-log-test",
		},
		"openErr": {
			fileErr:      true,
			errorOccurs:  true,
			errorMessage: "log.log",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			logDir = path.Join(os.TempDir(), "operator-log-test")
			logFilePath = path.Join(logDir, "log.log")

			require.NoError(t, os.RemoveAll(logDir))
			defer func() {
				_ = os.RemoveAll(logDir)
			}()

			if c.dirRO {
				require.NoError(t, ioutil.WriteFile(logDir, []byte(""), 0111))
			} else {
				require.NoError(t, os.MkdirAll(logDir, os.ModePerm))

				if c.fileErr {
					require.NoError(t, os.MkdirAll(logFilePath, 0111))
				}
			}

			f, err := LogFile()
			if c.errorOccurs {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errorMessage)
			} else {
				require.NoError(t, err)
				require.NoError(t, f.Close())
			}
		})
	}
}

func TestStartRotate(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		require.NoError(t, StartRotate("0 0 1 * * ?"))
	})

	t.Run("err", func(t *testing.T) {
		err := StartRotate("0 0 1 *?# * ?")
		require.Error(t, err)
		require.Equal(t, "Failed to parse int from *?#: strconv.Atoi: parsing \"*?#\": invalid syntax", err.Error())
	})
}

func Test_rotateLog(t *testing.T) {
	logDir = path.Join(os.TempDir(), "operator-log-test")
	logFilePath = path.Join(logDir, "log.log")

	tc := map[string]struct {
		rwFile    string
		wrongFile string
		lFile     string

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			rwFile: logFilePath,
			lFile:  logFilePath,
		},
		"readErr": {
			errorOccurs:  true,
			errorMessage: "log.log",
		},
		"writeErr": {
			rwFile:       logFilePath,
			wrongFile:    path.Join(logDir, fmt.Sprintf("%s.%s.log", logFilePrefix, time.Now().AddDate(0, 0, -1).Format("2006-01-02"))),
			errorOccurs:  true,
			errorMessage: "/operator.",
		},
		"truncErr": {
			wrongFile:    logFilePath,
			errorOccurs:  true,
			errorMessage: "log.log",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.RemoveAll(logDir))
			defer func() {
				_ = os.RemoveAll(logDir)
			}()
			require.NoError(t, os.MkdirAll(logDir, os.ModePerm))

			l := &test.FakeLogger{}
			logger = l

			if c.rwFile != "" {
				require.NoError(t, ioutil.WriteFile(c.rwFile, []byte(""), os.ModePerm))
			}
			if c.wrongFile != "" {
				require.NoError(t, os.MkdirAll(c.wrongFile, os.ModePerm))
			}

			if c.lFile != "" {
				var err error
				logFile, err = os.Open(c.lFile)
				require.NoError(t, err)
			} else {
				logFile = nil
			}

			rotateLog()

			if logFile != nil {
				_ = logFile.Close()
			}

			if c.errorOccurs {
				require.Len(t, l.Errors, 1)
				require.Contains(t, l.Errors[0].Error(), c.errorMessage)
			} else {
				require.Empty(t, l.Errors)
			}
		})
	}
}
