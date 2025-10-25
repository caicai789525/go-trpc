package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	client := NewFileTransferClient()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== 文件传输客户端 ===")
		fmt.Println("1. 上传文件")
		fmt.Println("2. 下载文件")
		fmt.Println("3. 列出文件")
		fmt.Println("4. 删除文件")
		fmt.Println("5. 退出")
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
			fmt.Println("再见!")
			return

		default:
			fmt.Println("无效选择，请重新输入")
		}
	}
}
