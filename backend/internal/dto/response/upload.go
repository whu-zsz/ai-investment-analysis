package response

type UploadRowError struct {
	RowNumber int    `json:"row_number"`
	Reason    string `json:"reason"`
}

type UploadResponse struct {
	FileID          int64            `json:"file_id"`
	FileName        string           `json:"file_name"`
	UploadStatus    string           `json:"upload_status"`
	RecordsTotal    int              `json:"records_total"`
	RecordsImported int              `json:"records_imported"`
	RecordsFailed   int              `json:"records_failed"`
	Errors          []UploadRowError `json:"errors"`
	Message         string           `json:"message"`
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
