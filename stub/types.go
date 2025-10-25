package file_transfer

// 所有数据结构的定义
type FileChunk struct {
	Filename    string `json:"filename,omitempty"`
	Content     []byte `json:"content,omitempty"`
	ChunkIndex  int64  `json:"chunk_index,omitempty"`
	TotalChunks int64  `json:"total_chunks,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
}

type FileRequest struct {
	Filename string `json:"filename,omitempty"`
}

type UploadResponse struct {
	Success  bool   `json:"success,omitempty"`
	Message  string `json:"message,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

type FileList struct {
	Files []string `json:"files,omitempty"`
}

type OperationResponse struct {
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

type Empty struct{}
