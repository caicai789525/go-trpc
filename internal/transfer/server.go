package transfer

import (
	"fmt"
	"net"
	"trpc-file/internal/ui"
)

func StartServer(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	ui.PrintInfo("Listening on port " + port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			ui.PrintError("Accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	var nameLen uint32
	fmt.Fscan(conn, &nameLen)
}
