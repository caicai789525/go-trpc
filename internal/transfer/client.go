package transfer

import (
	"encoding/binary"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"net"
	"os"
	"trpc-file/internal/ui"
)

func SendFile(host, port, filePath string) error {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, _ := file.Stat()
	nameBytes := []byte(info.Name())

	binary.Write(conn, binary.BigEndian, uint32(len(nameBytes)))
	conn.Write(nameBytes)
	binary.Write(conn, binary.BigEndian, uint64(info.Size()))

	ui.PrintInfo(fmt.Sprintf("Sending file: %s (%d bytes)", info.Name(), info.Size()))

	bar := progressbar.DefaultBytes(info.Size(), "Uploading")
	io.Copy(io.MultiWriter(conn, bar), file)
	ui.PrintInfo("Transfer complete ✅")
	return nil
}
