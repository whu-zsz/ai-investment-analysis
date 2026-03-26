package response

type UploadResponse struct {
	FileID          int64  `json:"file_id"`
	FileName        string `json:"file_name"`
	RecordsImported int    `json:"records_imported"`
	Message         string `json:"message"`
}

type UploadHistoryResponse struct {
	ID              uint64 `json:"id"`
	FileName        string `json:"file_name"`
	FileSize        int64  `json:"file_size"`
	FileType        string `json:"file_type"`
	UploadStatus    string `json:"upload_status"`
	RecordsImported int    `json:"records_imported"`
	UploadedAt      string `json:"uploaded_at"`
	ProcessedAt     string `json:"processed_at"`
}
