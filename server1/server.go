package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/server"
)

// 本地类型定义，避免依赖复杂的 stub
type FileChunk struct {
	Filename    string
	Content     []byte
	ChunkIndex  int64
	TotalChunks int64
	FileSize    int64
}

type FileRequest struct {
	Filename string
}

type UploadResponse struct {
	Success  bool
	Message  string
	FilePath string
	FileSize int64
}

type FileList struct {
	Files []string
}

type OperationResponse struct {
	Success bool
	Message string
}

type Empty struct{}

// 服务接口
type FileTransferService interface {
	UploadFile(server.Stream) error
	DownloadFile(context.Context, *FileRequest, server.Stream) error
	ListFiles(context.Context, *Empty) (*FileList, error)
	DeleteFile(context.Context, *FileRequest) (*OperationResponse, error)
}

// 服务实现
type FileTransferServiceImpl struct {
	fileMutex sync.Mutex
}

const (
	uploadDir   = "./uploads_server1"
	chunkSize   = 64 * 1024
	maxFileSize = 100 * 1024 * 1024
)

// 初始化
func init() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
}

// 实现接口方法
func (s *FileTransferServiceImpl) UploadFile(stream server.Stream) error {
	var file *os.File
	var receivedBytes int64
	var filename string

	for {
		chunk := &FileChunk{}
		if err := stream.RecvMsg(chunk); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if file == nil {
			filename = chunk.Filename
			if !isSafeFilename(filename) {
				return errs.New(400, "文件名不安全")
			}

			filePath := filepath.Join(uploadDir, filename)
			var err error
			file, err = os.Create(filePath)
			if err != nil {
				return errs.New(500, "创建文件失败: "+err.Error())
			}
			defer file.Close()
		}

		n, err := file.Write(chunk.Content)
		if err != nil {
			return errs.New(500, "写入文件失败: "+err.Error())
		}
		receivedBytes += int64(n)

		fmt.Printf("服务器1接收文件块 %d/%d, 大小: %d bytes\n",
			chunk.ChunkIndex+1, chunk.TotalChunks, len(chunk.Content))
	}

	resp := &UploadResponse{
		Success:  true,
		Message:  "文件上传成功",
		FilePath: filepath.Join(uploadDir, filename),
		FileSize: receivedBytes,
	}

	return stream.SendMsg(resp)
}

func (s *FileTransferServiceImpl) DownloadFile(ctx context.Context, req *FileRequest, stream server.Stream) error {
	if !isSafeFilename(req.Filename) {
		return errs.New(400, "文件名不安全")
	}

	filePath := filepath.Join(uploadDir, req.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		return errs.New(404, "文件不存在")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return errs.New(500, "获取文件信息失败")
	}

	fileSize := fileInfo.Size()
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	buffer := make([]byte, chunkSize)

	for i := int64(0); i < totalChunks; i++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return errs.New(500, "读取文件失败")
		}

		if n == 0 {
			break
		}

		chunk := &FileChunk{
			Filename:    req.Filename,
			Content:     buffer[:n],
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			FileSize:    fileSize,
		}

		if err := stream.SendMsg(chunk); err != nil {
			return err
		}

		fmt.Printf("服务器1发送文件块 %d/%d, 大小: %d bytes\n", i+1, totalChunks, n)
	}

	return nil
}

func (s *FileTransferServiceImpl) ListFiles(ctx context.Context, empty *Empty) (*FileList, error) {
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return nil, errs.New(500, "读取目录失败")
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return &FileList{Files: files}, nil
}

func (s *FileTransferServiceImpl) DeleteFile(ctx context.Context, req *FileRequest) (*OperationResponse, error) {
	s.fileMutex.Lock()
	defer s.fileMutex.Unlock()

	if !isSafeFilename(req.Filename) {
		return nil, errs.New(400, "文件名不安全")
	}

	filePath := filepath.Join(uploadDir, req.Filename)
	if err := os.Remove(filePath); err != nil {
		return &OperationResponse{
			Success: false,
			Message: "删除文件失败: " + err.Error(),
		}, nil
	}

	return &OperationResponse{
		Success: true,
		Message: "文件删除成功",
	}, nil
}

func isSafeFilename(filename string) bool {
	if filename == "" || strings.Contains(filename, "..") ||
		strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}
	return true
}
