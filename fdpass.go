package fd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func Send(fd, sendfd int) os.Error {
	fdlen := int(unsafe.Sizeof(sendfd))

	var cmsg syscall.Cmsghdr
	cmsg.Level = syscall.SOL_SOCKET
	cmsg.Type = syscall.SCM_RIGHTS
	cmsg.SetLen(cmsgLen(fdlen))

	end := endianess()

	cmsgbuf := bytes.NewBuffer(make([]byte, 0, cmsgLen(0)))
	err := binary.Write(cmsgbuf, end, &cmsg)
	if err != nil {
		return fmt.Errorf("fdpass: Send: %v", err)
	}
	sendfdbuf := bytes.NewBuffer(make([]byte, 0, fdlen))
	err = binary.Write(sendfdbuf, end, uint32(sendfd))
	if err != nil {
		return fmt.Errorf("fdpass: Send: %v", err)
	}
	cmsgbuf.ReadFrom(sendfdbuf)

	errno := syscall.Sendmsg(fd, nil, cmsgbuf.Bytes(), nil, 0)
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
	end := endianess()

	var cmsg syscall.Cmsghdr
	cmsgbuf := bytes.NewBuffer(buf[:cmsgLen(0)])
	err := binary.Read(cmsgbuf, end, &cmsg)
	if err != nil {
		return -1, fmt.Errorf("fdpass: Receive: %v", err)
	}
	if uint64(cmsg.Len) != uint64(cmsglen) {
		return -1, fmt.Errorf("fdpass: Receive: bad length %v", cmsg.Len)
	}
	if cmsg.Level != syscall.SOL_SOCKET {
		return -1, fmt.Errorf("fdpass: Receive: bad level %v", cmsg.Level)
	}
	if cmsg.Type != syscall.SCM_RIGHTS {
		return -1, fmt.Errorf("fdpass: Receive: bad type %v", cmsg.Type)
	}

	var sendfd uint32
	sendfdbuf := bytes.NewBuffer(buf[cmsgLen(0):cmsglen])
	err = binary.Read(sendfdbuf, end, &sendfd)
	if err != nil {
		return -1, fmt.Errorf("fdpass: Receive: %v", err)
	}

	return int(sendfd), nil
}

func cmsgLen(l int) int {
	return cmsgAlign(syscall.SizeofCmsghdr) + l
}

func cmsgAlign(l int) int {
	var dummy uint
	size := int(unsafe.Sizeof(dummy))
	return (l + size - 1) & ^(size - 1)
}

func endianess() binary.ByteOrder {
	var one uint16 = 1
	if *(*byte)(unsafe.Pointer(&one)) == 0 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}
