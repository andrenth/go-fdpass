package main

import (
	"fmt"
	"os"
	"syscall"
	"strconv"
	fdpass "github.com/andrenth/go-fdpass"
)

func readAndPrint(fd int, buf []byte) {
	_, errno := syscall.Read(fd, buf)
	if errno != 0 {
		fmt.Printf("readAndPrint: %v: %v\n", errno, syscall.Errstr(errno))
		os.Exit(1)
	}
	println(string(buf))
}

func getFdFromChild(pid, fd int) {
	var wstatus syscall.WaitStatus
	_, errno := syscall.Wait4(pid, &wstatus, 0, nil)
	if errno != 0 {
		fmt.Printf("getFdFromChild: Wait4: %v: %v\n", errno, syscall.Errstr(errno))
		os.Exit(1)
	}
	if wstatus.ExitStatus() != 0 {
		fmt.Printf("openfile exited(%v) with status %v\n", wstatus.Exited(), wstatus.ExitStatus())
		os.Exit(1)
	}
	var smallbuf [10]byte
	var largebuf [4096]byte
	readAndPrint(fd, smallbuf[:])
	recvfd, err := fdpass.Receive(fd)
	if err != nil {
		fmt.Printf("getFdFromChild: fdpass.Receive: %v\n", err)
		os.Exit(1)
	}
	readAndPrint(recvfd, largebuf[:])
	readAndPrint(fd, smallbuf[:])
}

func main() {
	fds, errno := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if errno != 0 {
		fmt.Printf("mycat: socketpair: %v: %v\n", errno, syscall.Errstr(errno))
		os.Exit(errno)
	}
	fd0, fd1 := fds[0], fds[1]

	syscall.CloseOnExec(fd0)
	bin := "./openfile/openfile"
	argv := []string{bin, strconv.Itoa(fd1), os.Args[1]}

	pid, errno := syscall.ForkExec(bin, argv, nil)
	if errno != 0 {
		fmt.Printf("getFdFromChild: ForkExec: %v\n", syscall.Errstr(errno))
		os.Exit(1)
	}

	syscall.Close(fd1)
	getFdFromChild(pid, fd0)
	syscall.Close(fd0)
}
