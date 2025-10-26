package transfer

import (
	"encoding/binary"
	"io"
	"net"
	"os"
)

func sendFileData(conn net.Conn, file *os.File, info os.FileInfo) error {
	name := []byte(info.Name())
	binary.Write(conn, binary.BigEndian, uint32(len(name)))
	conn.Write(name)
	binary.Write(conn, binary.BigEndian, uint64(info.Size()))
	_, err := io.Copy(conn, file)
	return err
}

func receiveFileData(conn net.Conn, dir string) (string, error) {
	var nameLen uint32
	if err := binary.Read(conn, binary.BigEndian, &nameLen); err != nil {
		return "", err
	}
	name := make([]byte, nameLen)
	io.ReadFull(conn, name)

	var size uint64
	binary.Read(conn, binary.BigEndian, &size)

	filePath := dir + "/" + string(name)
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	io.CopyN(f, conn, int64(size))
	return filePath, nil
}
