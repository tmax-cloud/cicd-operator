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

package test

import "github.com/go-logr/logr"

// FakeLogger is a logger for testing
type FakeLogger struct {
	Infos     []string
	Errors    []error
	ErrorMsgs []string
}

// Clear clears the logger
func (f *FakeLogger) Clear() {
	f.Infos = nil
	f.Errors = nil
	f.ErrorMsgs = nil
}

// Enabled returns true
func (f *FakeLogger) Enabled() bool { return true }

// Info logs info level
func (f *FakeLogger) Info(msg string, _ ...interface{}) {
	f.Infos = append(f.Infos, msg)
}

// Error logs error level
func (f *FakeLogger) Error(err error, msg string, _ ...interface{}) {
	f.Errors = append(f.Errors, err)
	f.ErrorMsgs = append(f.ErrorMsgs, msg)
}

// V returns the logger
func (f *FakeLogger) V(_ int) logr.Logger { return f }

// WithValues returns the logger
func (f *FakeLogger) WithValues(_ ...interface{}) logr.Logger { return f }

// WithName returns the logger
func (f *FakeLogger) WithName(_ string) logr.Logger { return f }
