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
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/envconfig"
)

var (
	version string = "0.0.0"
	commit  string = ""
)

var goenv struct {
	Debug     bool   `default:"false"`
	LogPrefix string `default:"nafder"`
}

var checkInterval int = 30 //sec
var cancelList map[string]context.CancelFunc
var cancelMutex sync.Mutex

var logInfo nafderLogger
var logWarn nafderLogger
var logErr nafderLogger
var logDebug nafderLogger
var logApp nafderLogger

var (
	flagCopyTo  string
	flagTime    bool
	flagDebug   bool
	flagHelp    bool
	flagVersion bool
)

func showHelp() {
	var help = `
Nafder is a tool to identifying logs for container which has multi application.
https://github.com/t-matsuo/nafder

Usage:
  $ nafder [options...] TargetDir

Options:
   -c --copy    : copy logs to specified file
   -t --time    : output timestamp
   -d --debug   : Print debug messages
   --version    : Show version number
   -h --help    : Show help

Environment Variables:
   NAFDER_DEBUG=true : Print debug messages (=--debug option)
`
	fmt.Println(help)
}

func init() {
	logInfo.main = log.New(os.Stdout, "nafder INFO ", log.Lmsgprefix)
	logErr.main = log.New(os.Stderr, "nafder ERROR ", log.Lmsgprefix)
	logWarn.main = log.New(os.Stderr, "nafder WARN ", log.Lmsgprefix)
	logDebug.main = log.New(ioutil.Discard, "nafder DEBUG ", log.Lmsgprefix)
	logApp.main = log.New(os.Stdout, "", log.Lmsgprefix)
}

func setupFlags() {
	flagDebug = true
	flag.StringVar(&flagCopyTo, "c", "", "copy logs to specified file")
	flag.StringVar(&flagCopyTo, "copy", "", "copy logs to specified file(=-c)")
	flag.BoolVar(&flagTime, "t", false, "output timestamp")
	flag.BoolVar(&flagTime, "time", false, "output timestamp(=-t)")
	flag.BoolVar(&flagDebug, "d", false, "Print debug messages")
	flag.BoolVar(&flagDebug, "debug", false, "Print debug messages (=-d)")
	flag.BoolVar(&flagHelp, "h", false, "Show help")
	flag.BoolVar(&flagHelp, "help", false, "Show help (=-h)")
	flag.BoolVar(&flagVersion, "version", false, "Show version")
	flag.Parse()
}

func openSubLogger() *os.File {
	if flagCopyTo == "" {
		return nil
	}
	logInfo.Println("Copy to", flagCopyTo)

	f, err := os.OpenFile(flagCopyTo, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		logErr.Println("Cannot open", flagCopyTo)
		return nil
	}
	logInfo.sub = log.New(f, "nafder INFO ", log.Lmsgprefix)
	logErr.sub = log.New(f, "nafder ERROR ", log.Lmsgprefix)
	logWarn.sub = log.New(f, "nafder WARN ", log.Lmsgprefix)
	logDebug.sub = log.New(ioutil.Discard, "nafder DEBUG ", log.Lmsgprefix)
	logApp.sub = log.New(f, "", log.Lmsgprefix)
	return f
}

func handleEnv(f *os.File) {
	if err := envconfig.Process("NAFDER", &goenv); err != nil {
		logErr.Fatalf("Failed to process env: %s", err)
		os.Exit(1)
	}
	if flagTime {
		logInfo.EnableTimestamp()
		logWarn.EnableTimestamp()
		logErr.EnableTimestamp()
		logDebug.EnableTimestamp()
		logApp.EnableTimestamp()
	}

	// setup log outputs
	if goenv.Debug || flagDebug {
		logDebug.EnableOutput(f)
		logInfo.EnableDebug()
		logWarn.EnableDebug()
		logErr.EnableDebug()
		logDebug.EnableDebug()
		logApp.EnableDebug()
		logInfo.Println("Debug message is enabled")
	}

	logInfo.SetPrefix(goenv.LogPrefix + " INFO ")
	logWarn.SetPrefix(goenv.LogPrefix + " WARN ")
	logErr.SetPrefix(goenv.LogPrefix + " ERROR ")
	logDebug.SetPrefix(goenv.LogPrefix + " DEBUG ")
}

func getTargetDir() string {
	var targetDir string

	if args := flag.Args(); len(args) > 0 {
		targetDir = args[0]
	} else {
		logErr.Println("TargetDir is not specified")
		showHelp()
		os.Exit(1)
	}

	if f, err := os.Stat(targetDir); os.IsNotExist(err) || !f.IsDir() {
		logErr.Println("TargetDir not found")
		os.Exit(1)
	}
	return targetDir
}

func isPipe(pipe string) bool {
	if p, err := os.Stat(pipe); os.IsNotExist(err) || (p.Mode()&os.ModeNamedPipe) <= 0 {
		logDebug.Println(pipe, "is not named pipe")
		return false
	}
	return true
}

func addPrefix(ctx context.Context, pipe string) {
	p, err := os.OpenFile(pipe, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe)
	if err != nil {
		logErr.Println("Open named pipe file error:", err)
		checkAndDelCancelListAndCancel(pipe)
		return
	}
	defer logInfo.Println("Closing", pipe)
	defer p.Close()

	prefix := path.Base(pipe)
	reader := bufio.NewReader(p)
	logInfo.Println("Reading", pipe)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if _, err := os.Stat(pipe); os.IsNotExist(err) {
				logInfo.Println(pipe, "is deleted")
				checkAndDelCancelListAndCancel(pipe)
				return
			}
			line, err := reader.ReadBytes('\n')
			if len(line) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if err == nil {
				logApp.Print(prefix + " " + string(line))
			} else {
				_ = line
				fmt.Printf("READ err: %v\n", err)
				break
			}
		}
	}
}

func checkAndAddCancelList(pipe string, cancel context.CancelFunc) bool {
	defer cancelMutex.Unlock()
	cancelMutex.Lock()
	if _, haskey := cancelList[pipe]; haskey {
		logDebug.Println(pipe, "is already added")
		return false
	} else {
		logDebug.Println("Adding", pipe, "into list")
		cancelList[pipe] = cancel
		return true
	}
}

func checkAndDelCancelListAndCancel(pipe string) bool {
	defer cancelMutex.Unlock()
	logDebug.Println("Deleting and Canceling", pipe)
	cancelMutex.Lock()
	if _, haskey := cancelList[pipe]; haskey {
		delete(cancelList, pipe)
		if cancel := cancelList[pipe]; cancel != nil {
			cancel()
		} else {
			logDebug.Println(pipe, "is already canceled")
		}
		return true
	} else {
		logDebug.Println(pipe, "already deleted")
		return false
	}
}

func parseDir(dir string) {
	logDebug.Println("Parsing " + dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if isPipe(path.Join(dir, file.Name())) {
			logDebug.Println(dir, "has pipe", file.Name())
			ctx, cancel := context.WithCancel(context.Background())
			if checkAndAddCancelList(path.Join(dir, file.Name()), cancel) {
				go addPrefix(ctx, path.Join(dir, file.Name()))
			}
		}
	}
}

func handleInotify(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			//logDebug.Println("fsnotify event:", event)
			if event.Op == fsnotify.Remove {
				checkAndDelCancelListAndCancel(string(event.Name))
			}
			if event.Op == fsnotify.Create {
				if isPipe(string(event.Name)) {
					logInfo.Println("named pipe", string(event.Name), "is created")
					ctx, cancel := context.WithCancel(context.Background())
					if checkAndAddCancelList(string(event.Name), cancel) {
						go addPrefix(ctx, string(event.Name))
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logErr.Fatalln("fsnotify error:", err)
		}
	}

}

func mainLoop(dir string) {
	logDebug.Println("start mainloop")
	for {
		time.Sleep(time.Millisecond * time.Duration(checkInterval) * 1000)
		parseDir(dir)
	}
}

func main() {
	setupFlags()
	f := openSubLogger()
	handleEnv(f)

	if flagVersion {
		fmt.Println("nafder " + version)
		fmt.Println("Commit " + commit)
		fmt.Println("Source https://github.com/t-matsuo/nafder")
		os.Exit(0)
	}

	if flagHelp {
		showHelp()
		os.Exit(0)
	}

	cancelList = make(map[string]context.CancelFunc, 2)

	targetDir := getTargetDir()
	logDebug.Println("Target Directory is", targetDir)
	parseDir(targetDir)
	go handleInotify(targetDir)
	mainLoop(targetDir)
}
