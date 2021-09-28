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

package controllers

import "github.com/go-logr/logr"

type fakeLogger struct {
	info     []string
	error    []error
	errorMsg []string
}

func (f *fakeLogger) Clear() {
	f.info = nil
	f.error = nil
	f.errorMsg = nil
}

func (f *fakeLogger) Info(msg string, _ ...interface{}) {
	f.info = append(f.info, msg)
}
func (f *fakeLogger) Enabled() bool { return true }
func (f *fakeLogger) Error(err error, msg string, _ ...interface{}) {
	f.error = append(f.error, err)
	f.errorMsg = append(f.errorMsg, msg)
}
func (f *fakeLogger) V(_ int) logr.InfoLogger                 { return f }
func (f *fakeLogger) WithValues(_ ...interface{}) logr.Logger { return f }
func (f *fakeLogger) WithName(_ string) logr.Logger           { return f }
