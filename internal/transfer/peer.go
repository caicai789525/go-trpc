package transfer

import (
	"encoding/binary"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"net"
	"os"
	"path/filepath"
	"trpc-file/internal/config"
	"trpc-file/internal/ui"
)

func StartPeer(port string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		ui.PrintError("Failed to load config.json:", err)
		return
	}

	// 启动接收端
	go func() {
		ln, _ := net.Listen("tcp", ":"+cfg.Port)
		ui.PrintInfo("Peer listening on " + cfg.SelfIP + ":" + cfg.Port)
		for {
			conn, err := ln.Accept()
			if err != nil {
				ui.PrintError("Accept error:", err)
				continue
			}
			go handleIncoming(conn)
		}
	}()

	// 打印本机和对端信息
	ui.PrintInfo("This server IP:", cfg.SelfIP)
	ui.PrintInfo("Peer server IP:", cfg.PeerIP)

	select {}
}

func handleIncoming(conn net.Conn) {
	defer conn.Close()

	var nameLen uint32
	binary.Read(conn, binary.BigEndian, &nameLen)
	name := make([]byte, nameLen)
	io.ReadFull(conn, name)

	var fileSize uint64
	binary.Read(conn, binary.BigEndian, &fileSize)

	savePath := filepath.Join("./received_" + string(name))
	file, err := os.Create(savePath)
	if err != nil {
		ui.PrintError("Create file error:", err)
		return
	}
	defer file.Close()

	ui.PrintInfo(fmt.Sprintf("Receiving file: %s (%d bytes)", name, fileSize))
	bar := progressbar.DefaultBytes(int64(fileSize), "Receiving")
	io.Copy(io.MultiWriter(file, bar), conn)
	ui.PrintInfo("Saved to:", savePath)
}
