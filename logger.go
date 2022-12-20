package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func logger() func() {
	var root, _ = os.UserConfigDir()
	var configuration = filepath.Join(root, "senna")
	os.Mkdir(configuration, os.ModePerm)
	var path = filepath.Join(configuration, "log.txt")

	file, _ := os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	stdout := os.Stdout
	multiwriter := io.MultiWriter(stdout, file)
	read, write, _ := os.Pipe()

	os.Stdout = write
	os.Stderr = write

	log.SetOutput(multiwriter)
	exit := make(chan bool)

	go func() {
		_, _ = io.Copy(multiwriter, read)
		exit <- true
	}()

	return func() {
		_ = write.Close()
		<-exit
		_ = file.Close()
	}
}
