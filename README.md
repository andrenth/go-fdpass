Go-FDPass
=========

This is a Go package that allows file descriptor passing between processes.
This is usually done through a socket pair.

Functions
---------

    func Send(fd, sendfd int) os.Error
    func Receive(fd int) (int, os.Error)
