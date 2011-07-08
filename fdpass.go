package fd

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func Send(fd, sendfd int) os.Error {
	cmsglen := cmsgLen(int(unsafe.Sizeof(sendfd)))

	buf := make([]byte, cmsglen)
	cmsg := (*syscall.Cmsghdr)(unsafe.Pointer(&buf[0]))
	cmsg.Level = syscall.SOL_SOCKET
	cmsg.Type = syscall.SCM_RIGHTS
	cmsg.SetLen(cmsglen)
	*(*int)(unsafe.Pointer(&buf[cmsgLen(0)])) = sendfd

	errno := syscall.Sendmsg(fd, nil, buf, nil, 0)
	if errno != 0 {
		return fmt.Errorf("fdpass: Send: %v", os.Errno(errno))
	}
	return nil
}

func Receive(fd int) (int, os.Error) {
	cmsglen := cmsgLen(int(unsafe.Sizeof(fd)))
	buf := make([]byte, cmsglen)
	_, _, _, _, errno := syscall.Recvmsg(fd, nil, buf, 0)
	if errno != 0 {
		return -1, fmt.Errorf("fdpass: Receive: %v", os.Errno(errno))
	}

	cmsg := (*syscall.Cmsghdr)(unsafe.Pointer(&buf[0]))
	if uint64(cmsg.Len) != uint64(cmsglen) {
		return -1, fmt.Errorf("fdpass: Receive: bad length %v", cmsg.Len)
	}
	if cmsg.Level != syscall.SOL_SOCKET {
		return -1, fmt.Errorf("fdpass: Receive: bad level %v", cmsg.Level)
	}
	if cmsg.Type != syscall.SCM_RIGHTS {
		return -1, fmt.Errorf("fdpass: Receive: bad type %v", cmsg.Type)
	}

	sendfd := *(*int)(unsafe.Pointer(&buf[cmsgLen(0)]))
	return sendfd, nil
}

func cmsgLen(l int) int {
	return cmsgAlign(syscall.SizeofCmsghdr) + l
}

func cmsgAlign(l int) int {
	var dummy uint
	size := int(unsafe.Sizeof(dummy))
	return (l + size - 1) & ^(size - 1)
}
