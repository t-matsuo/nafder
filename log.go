/*
Copyright 2021 MATSUO Takatoshi

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

package main

import (
	"log"
	"os"
	"runtime"
	"strconv"
)

type nafderLogger struct {
	main    *log.Logger
	sub     *log.Logger
	isDebug bool
	prefix  string
}

func (l *nafderLogger) Print(v ...interface{}) {
	var file string
	var line int
	if l.isDebug {
		_, file, line, _ = runtime.Caller(1)
		l.main.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
	}
	l.main.Print(v...)
	if l.sub != nil {
		if l.isDebug {
			l.sub.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
		}
		l.sub.Print(v...)
	}
}

func (l *nafderLogger) Println(v ...interface{}) {
	var file string
	var line int
	if l.isDebug {
		_, file, line, _ = runtime.Caller(1)
		l.main.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
	}
	l.main.Println(v...)
	if l.sub != nil {
		if l.isDebug {
			l.sub.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
		}
		l.sub.Println(v...)
	}
}

func (l *nafderLogger) Fatal(v ...interface{}) {
	var file string
	var line int
	if l.isDebug {
		_, file, line, _ = runtime.Caller(1)
		l.main.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
	}
	l.main.Fatal(v...)
	if l.sub != nil {
		if l.isDebug {
			l.sub.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
		}
		l.sub.Fatal(v...)
	}
}

func (l *nafderLogger) Fatalln(v ...interface{}) {
	var file string
	var line int
	if l.isDebug {
		_, file, line, _ = runtime.Caller(1)
		l.main.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
	}
	l.main.Fatalln(v...)
	if l.sub != nil {
		if l.isDebug {
			l.sub.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
		}
		l.sub.Fatalln(v...)
	}
}

func (l *nafderLogger) Fatalf(format string, v ...interface{}) {
	var file string
	var line int
	if l.isDebug {
		_, file, line, _ = runtime.Caller(1)
		l.main.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
	}
	l.main.Fatalf(format, v...)
	if l.sub != nil {
		if l.isDebug {
			l.sub.SetPrefix(file + ":" + strconv.Itoa(line) + ": " + l.prefix)
		}
		l.sub.Fatalf(format, v...)
	}
}

func (l *nafderLogger) SetFlags(flag int) {
	l.main.SetFlags(flag)
	if l.sub != nil {
		l.sub.SetFlags(flag)
	}
}

func (l *nafderLogger) SetPrefix(prefix string) {
	l.prefix = prefix
	l.main.SetPrefix(prefix)
	if l.sub != nil {
		l.sub.SetPrefix(prefix)
	}
}

func (l *nafderLogger) EnableOutput(f *os.File) {
	l.main.SetOutput(os.Stderr)
	if l.sub != nil {
		l.sub.SetOutput(f)
	}
}

func (l *nafderLogger) EnableDebug() {
	l.isDebug = true
}

func (l *nafderLogger) EnableTimestamp() {
	l.main.SetFlags(log.LstdFlags | log.Lmsgprefix)
	if l.sub != nil {
		l.sub.SetFlags(log.LstdFlags | log.Lmsgprefix)
	}
}
