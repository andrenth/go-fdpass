package main

import (
	"os"
	"strconv"
	"syscall"
	fdpass "github.com/andrenth/go-fdpass"
)

func write(fd int, s string) {
	b := []byte(s)
	n, errno := syscall.Write(fd, b)
	if errno != 0 {
		os.Exit(7)
	}
	if n != len(s) {
		os.Exit(8)
	}
}

func main() {
	fd, err := strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(10)
	}
	defer syscall.Close(fd)
	path := os.Args[2]
	fi, err := os.Stat(path)
	if err != nil {
		os.Exit(20)
	}
	if !fi.IsRegular() {
		os.Exit(30)
	}
	f, err := os.Open(path)
	if err != nil {
		os.Exit(40)
	}
	defer f.Close()
	write(fd, "1234567890")
	err = fdpass.Send(fd, f.Fd())
	if err != nil {
		os.Exit(50)
	}
	write(fd, "0987654321")
	os.Exit(0)
}
