# Go commons

[![GoDoc](https://godoc.org/github.com/nickbruun/gocommons?status.svg)](https://godoc.org/github.com/nickbruun/gocommons) [![Build status](https://travis-ci.org/nickbruun/gocommons.svg?branch=master)](https://travis-ci.org/nickbruun/gocommons)

Common libraries for the type of Go applications I toy with and build.


## Contents

The libraries are a mixture of foundation, utility and system libraries for various purposes.


### Foundation

**logging**

A thin wrapper library around [Simon Eskildsen's](http://sirupsen.com/) excellent [logrus](https://github.com/Sirupsen/logrus) that provides caller annotations to make it easier to figure out where log messages are coming from:

```Go
package main

import (
	logrus "github.com/Sirupsen/logrus"
	log "github.com/nickbruun/gocommons/logging"
	"sync"
)

var wg sync.WaitGroup

func MyFunc() {
	log.Debug("Debug time!")

	wg.Add(1)
	go func() {
		log.Info("Go-routines are presented understandable as well")
		wg.Done()
	}()
	wg.Wait()
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	log.Info("Hola, mundo")
	MyFunc()
}
```

Results in:

    INFO[0000] Hola, mundo                                       file=/Users/nick/Code/gocommonsdev/src/github.com/nickbruun/gocommons/play/main.go func=main.main line=25
    DEBU[0000] Debug time!                                       file=/Users/nick/Code/gocommonsdev/src/github.com/nickbruun/gocommons/play/main.go func=main.MyFunc line=12
    INFO[0000] Go-routines are presented understandable as well  file=/Users/nick/Code/gocommonsdev/src/github.com/nickbruun/gocommons/play/main.go func=main.func·001 line=16
