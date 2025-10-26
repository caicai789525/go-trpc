package main

import (
	"flag"
	"fmt"
	"os"
	"trpc-file/internal/config"

	"trpc-file/internal/transfer"
	"trpc-file/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		ui.ShowLogo()
		fmt.Println("[Usage]")
		fmt.Println("  trpc-file send --host <ip> --port <port> --file <path>")
		fmt.Println("  trpc-file recv --port <port>")
		fmt.Println("  trpc-file peer --port <port>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "send":
		sendCmd()
	case "recv":
		recvCmd()
	case "peer":
		peerCmd()
	case "autosend":
		autoSendCmd()
	default:
		ui.PrintError("Unknown command:", cmd)
	}
}

func sendCmd() {
	host := flag.String("host", "", "target host")
	port := flag.String("port", "8080", "target port")
	file := flag.String("file", "", "file path to send")
	flag.CommandLine.Parse(os.Args[2:])

	if *host == "" || *file == "" {
		ui.PrintError("Usage: trpc-file send --host <ip> --port <port> --file <path>")
		os.Exit(1)
	}

	ui.ShowLogo()
	ui.PrintInfo("Connecting to " + *host + ":" + *port + " ...")
	if err := transfer.SendFile(*host, *port, *file); err != nil {
		ui.PrintError("Send failed:", err)
	}
}

func autoSendCmd() {
	file := flag.String("file", "", "file to send automatically")
	flag.CommandLine.Parse(os.Args[2:])

	cfg, err := config.LoadConfig()
	if err != nil {
		ui.PrintError("Failed to load config:", err)
		os.Exit(1)
	}

	if *file == "" {
		ui.PrintError("Missing parameter: --file")
		os.Exit(1)
	}

	ui.ShowLogo()
	ui.PrintInfo(fmt.Sprintf("Auto sending %s to %s:%s", *file, cfg.PeerIP, cfg.Port))
	if err := transfer.SendFile(cfg.PeerIP, cfg.Port, *file); err != nil {
		ui.PrintError("Send failed:", err)
	}
}

func recvCmd() {
	port := flag.String("port", "8080", "port to listen on")
	flag.CommandLine.Parse(os.Args[2:])
	ui.ShowLogo()
	if err := transfer.StartServer(*port); err != nil {
		ui.PrintError("Server error:", err)
	}
}

func peerCmd() {
	port := flag.String("port", "8080", "port to listen on")
	flag.CommandLine.Parse(os.Args[2:])
	ui.ShowLogo()
	ui.PrintInfo("Starting peer mode on port " + *port)
	transfer.StartPeer(*port)
}
