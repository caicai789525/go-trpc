package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// 选择服务器
	fmt.Println("选择要连接的服务器:")
	fmt.Println("1. 服务器1 (39.96.188.155:8000)")
	fmt.Println("2. 服务器2 (172.25.112.0:8001)")
	fmt.Println("3. 自定义服务器地址")
	fmt.Print("请选择: ")

	scanner.Scan()
	serverChoice := strings.TrimSpace(scanner.Text())

	var serverAddr string
	switch serverChoice {
	case "1":
		serverAddr = "39.96.188.155:8000"
	case "2":
		serverAddr = "172.25.112.0:8001"
	case "3":
		fmt.Print("请输入服务器地址 (格式: IP:端口): ")
		scanner.Scan()
		serverAddr = strings.TrimSpace(scanner.Text())
	default:
		fmt.Println("无效选择，默认连接到服务器1")
		serverAddr = "39.96.188.155:8000"
	}

	client := NewFileTransferClient(serverAddr)

	for {
		fmt.Printf("\n=== 文件传输客户端 (服务器: %s) ===\n", serverAddr)
		fmt.Println("1. 上传文件")
		fmt.Println("2. 下载文件")
		fmt.Println("3. 列出文件")
		fmt.Println("4. 删除文件")
		fmt.Println("5. 服务器间同步文件")
		fmt.Println("6. 切换服务器")
		fmt.Println("7. 退出")
		fmt.Print("请选择操作: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			fmt.Print("请输入要上传的文件路径: ")
			scanner.Scan()
			filePath := strings.TrimSpace(scanner.Text())
			if err := client.UploadFile(filePath); err != nil {
				log.Printf("上传失败: %v", err)
			}

		case "2":
			fmt.Print("请输入要下载的文件名: ")
			scanner.Scan()
			filename := strings.TrimSpace(scanner.Text())
			fmt.Print("请输入保存路径: ")
			scanner.Scan()
			savePath := strings.TrimSpace(scanner.Text())
			if err := client.DownloadFile(filename, savePath); err != nil {
				log.Printf("下载失败: %v", err)
			}

		case "3":
			if err := client.ListFiles(); err != nil {
				log.Printf("获取文件列表失败: %v", err)
			}

		case "4":
			fmt.Print("请输入要删除的文件名: ")
			scanner.Scan()
			filename := strings.TrimSpace(scanner.Text())
			if err := client.DeleteFile(filename); err != nil {
				log.Printf("删除失败: %v", err)
			}

		case "5":
			fmt.Print("请输入要同步的文件名: ")
			scanner.Scan()
			filename := strings.TrimSpace(scanner.Text())
			fmt.Print("请输入目标服务器地址 (格式: IP:端口): ")
			scanner.Scan()
			targetServer := strings.TrimSpace(scanner.Text())
			if err := client.SyncBetweenServers(filename, targetServer); err != nil {
				log.Printf("同步失败: %v", err)
			}

		case "6":
			fmt.Println("选择要连接的服务器:")
			fmt.Println("1. 服务器1 (39.96.188.155:8000)")
			fmt.Println("2. 服务器2 (172.25.112.0:8001)")
			fmt.Println("3. 自定义服务器地址")
			fmt.Print("请选择: ")
			scanner.Scan()
			newChoice := strings.TrimSpace(scanner.Text())
			switch newChoice {
			case "1":
				serverAddr = "39.96.188.155:8000"
			case "2":
				serverAddr = "172.25.112.0:8001"
			case "3":
				fmt.Print("请输入服务器地址 (格式: IP:端口): ")
				scanner.Scan()
				serverAddr = strings.TrimSpace(scanner.Text())
			default:
				fmt.Println("无效选择，保持当前服务器")
			}
			client = NewFileTransferClient(serverAddr)
			fmt.Printf("已切换到服务器: %s\n", serverAddr)

		case "7":
			fmt.Println("再见!")
			return

		default:
			fmt.Println("无效选择，请重新输入")
		}
	}
}
