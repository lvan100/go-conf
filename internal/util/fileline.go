// Copyright 2024 github.com/lvan100
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"runtime"
	"sync"
)

var frameMap sync.Map

func fileLine() (string, int) {
	rpc := make([]uintptr, 1)
	runtime.Callers(3, rpc[:])
	pc := rpc[0]
	if v, ok := frameMap.Load(pc); ok {
		e := v.(*runtime.Frame)
		return e.File, e.Line
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	frameMap.Store(pc, &frame)
	return frame.File, frame.Line
}

// FileLine returns the file name and line of the call point.
// In reality FileLine here costs less time than debug.Stack.
func FileLine() string {
	file, line := fileLine()
	return fmt.Sprintf("%s:%d", file, line)
}
