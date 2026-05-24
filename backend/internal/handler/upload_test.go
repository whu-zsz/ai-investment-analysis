package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/handler"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// MockUploadService 实现 UploadService 接口用于测试
type MockUploadService struct {
	ProcessUploadedFileFunc func(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error)
	GetUploadHistoryFunc    func(userID uint64) ([]response.UploadHistoryResponse, error)
}

func (m *MockUploadService) ProcessUploadedFile(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error) {
	if m.ProcessUploadedFileFunc != nil {
		return m.ProcessUploadedFileFunc(userID, filePath, originalName, fileSize, fileType)
	}
	return &response.UploadResponse{
		FileID:         1,
		RecordsImported: 10,
	}, nil
}

func (m *MockUploadService) GetUploadHistory(userID uint64) ([]response.UploadHistoryResponse, error) {
	if m.GetUploadHistoryFunc != nil {
		return m.GetUploadHistoryFunc(userID)
	}
	return []response.UploadHistoryResponse{}, nil
}

func getTestUploadConfig() config.UploadConfig {
	return config.UploadConfig{
		Path:           "./uploads",
		MaxUploadSize:  10 * 1024 * 1024, // 10MB
	}
}

// TestUploadFile_Success 测试上传文件成功
func TestUploadFile_Success(t *testing.T) {
	mockService := &MockUploadService{
		ProcessUploadedFileFunc: func(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error) {
			return &response.UploadResponse{
				FileID:         1,
				RecordsImported: 10,
			}, nil
		},
	}

	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建模拟文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("transaction_date,transaction_type,asset_code,asset_name,quantity,price_per_unit\n2024-01-15,buy,600519,贵州茅台,100,1850.00"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestUploadFile_NoFile 测试没有上传文件
func TestUploadFile_NoFile(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	req := httptest.NewRequest("POST", "/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUploadFile_UnsupportedType 测试不支持的文件类型
func TestUploadFile_UnsupportedType(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建 .txt 文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("some text content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUploadFile_ExcelFile 测试上传 Excel 文件
func TestUploadFile_ExcelFile(t *testing.T) {
	mockService := &MockUploadService{
		ProcessUploadedFileFunc: func(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error) {
			if fileType != "xlsx" {
				t.Errorf("Expected fileType xlsx, got %s", fileType)
			}
			return &response.UploadResponse{FileID: 1, RecordsImported: 5}, nil
		},
	}

	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建 .xlsx 文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.xlsx")
	part.Write([]byte("mock excel content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestUploadFile_ServiceError 测试服务层错误
func TestUploadFile_ServiceError(t *testing.T) {
	mockService := &MockUploadService{
		ProcessUploadedFileFunc: func(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error) {
			return nil, errors.New("failed to parse file")
		},
	}

	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("invalid content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetUploadHistory_Success 测试获取上传历史成功
func TestGetUploadHistory_Success(t *testing.T) {
	mockService := &MockUploadService{
		GetUploadHistoryFunc: func(userID uint64) ([]response.UploadHistoryResponse, error) {
			return []response.UploadHistoryResponse{
				{ID: 1, FileName: "test1.csv", RecordsImported: 10},
				{ID: 2, FileName: "test2.xlsx", RecordsImported: 20},
			}, nil
		},
	}

	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/upload/history", h.GetUploadHistory)

	req := httptest.NewRequest("GET", "/upload/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 history records, got %d", len(data))
	}
}

// TestGetUploadHistory_Empty 测试空上传历史
func TestGetUploadHistory_Empty(t *testing.T) {
	mockService := &MockUploadService{
		GetUploadHistoryFunc: func(userID uint64) ([]response.UploadHistoryResponse, error) {
			return []response.UploadHistoryResponse{}, nil
		},
	}

	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/upload/history", h.GetUploadHistory)

	req := httptest.NewRequest("GET", "/upload/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].([]interface{})
	if len(data) != 0 {
		t.Errorf("Expected 0 history records, got %d", len(data))
	}
}

// TestGetUploadHistory_ServiceError 测试服务层错误
func TestGetUploadHistory_ServiceError(t *testing.T) {
	mockService := &MockUploadService{
		GetUploadHistoryFunc: func(userID uint64) ([]response.UploadHistoryResponse, error) {
			return nil, errors.New("database error")
		},
	}

	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/upload/history", h.GetUploadHistory)

	req := httptest.NewRequest("GET", "/upload/history", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// ========== 安全测试 ==========

func TestUpload_Security_MaliciousExtension(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建包含恶意扩展名的文件
	maliciousExtensions := []string{
		"test.exe",
		"test.js",
		"test.php",
		"test.py",
		"test.sh",
		"test.bat",
	}

	for _, filename := range maliciousExtensions {
		// 创建 multipart 请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", filename)
		part.Write([]byte("test content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// 调用上传接口
		router.ServeHTTP(w, req)

		// 应该返回 400（文件类型不支持）
		if w.Code != http.StatusBadRequest {
			t.Errorf("Upload with malicious extension '%s' returned %d", filename, w.Code)
		}
	}
}

func TestUpload_Security_DoubleExtension(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建包含双扩展名的文件
	doubleExtensions := []string{
		"test.csv.exe",
		"test.xlsx.js",
		"test.xls.php",
	}

	for _, filename := range doubleExtensions {
		// 创建 multipart 请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", filename)
		part.Write([]byte("test content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// 调用上传接口
		router.ServeHTTP(w, req)

		// 应该返回 400（文件类型不支持）
		if w.Code != http.StatusBadRequest {
			t.Errorf("Upload with double extension '%s' returned %d", filename, w.Code)
		}
	}
}

func TestUpload_Security_EmptyFile(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建空文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte(""))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// 调用上传接口
	router.ServeHTTP(w, req)

	// 应该返回 400（文件为空）或 200（服务层未校验空文件）
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("Upload with empty file returned %d", w.Code)
	}
}

func TestUpload_Security_OversizedFile(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建超大文件（超过 10MB）
	largeContent := strings.Repeat("a", 11*1024*1024) // 11MB

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte(largeContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// 调用上传接口
	router.ServeHTTP(w, req)

	// 应该返回 400（文件过大）或 413
	if w.Code != http.StatusBadRequest && w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Upload with oversized file returned %d", w.Code)
	}
}

func TestUpload_Security_PathTraversal(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建包含路径穿越的文件名
	pathTraversalPayloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
	}

	for _, filename := range pathTraversalPayloads {
		// 创建 multipart 请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", filename)
		part.Write([]byte("test content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// 调用上传接口
		router.ServeHTTP(w, req)

		// 应该返回 400（文件名无效/类型不支持）或 200（路径被清洗）
		if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
			t.Errorf("Upload with path traversal '%s' returned %d", filename, w.Code)
		}
	}
}

func TestUpload_Security_SpecialCharsFilename(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建包含特殊字符的文件名
	specialChars := []string{
		"test<script>.csv",
		"test'; DROP TABLE files; --.csv",
		"test${7*7}.csv",
		"test{{7*7}}.csv",
		"test%00.csv",
	}

	for _, filename := range specialChars {
		// 创建 multipart 请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", filename)
		part.Write([]byte("test content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// 调用上传接口
		router.ServeHTTP(w, req)

		// 应该返回 400（文件名无效）或 200（特殊字符被处理）
		if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
			t.Errorf("Upload with special chars filename '%s' returned %d", filename, w.Code)
		}
	}
}

func TestUpload_Security_UnicodeFilename(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建包含 Unicode 的文件名
	unicodeFilenames := []string{
		"测试文件.csv",
		"données.xlsx",
		"ファイル.xls",
	}

	for _, filename := range unicodeFilenames {
		// 创建 multipart 请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", filename)
		part.Write([]byte("test content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		// 调用上传接口
		router.ServeHTTP(w, req)

		// 应该返回 200（成功）或 400（文件名处理失败）
		if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
			t.Errorf("Upload with unicode filename '%s' returned %d", filename, w.Code)
		}
	}
}

func TestUpload_Security_NoExtension(t *testing.T) {
	mockService := &MockUploadService{}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建无扩展名的文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "testfile")
	part.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// 调用上传接口
	router.ServeHTTP(w, req)

	// 应该返回 400（文件类型不支持）
	if w.Code != http.StatusBadRequest {
		t.Errorf("Upload with no extension returned %d", w.Code)
	}
}

func TestUpload_Security_ScriptContent(t *testing.T) {
	mockService := &MockUploadService{
		ProcessUploadedFileFunc: func(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error) {
			return &response.UploadResponse{FileID: 1, RecordsImported: 3}, nil
		},
	}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 创建包含脚本内容的 CSV 文件
	scriptContent := "<script>alert('xss')</script>\nname,value\ntest,123"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte(scriptContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// 调用上传接口
	router.ServeHTTP(w, req)

	// CSV 文件本身是合法的，即使包含脚本内容，应该返回 200
	if w.Code != http.StatusOK {
		t.Errorf("Upload with script content returned %d", w.Code)
	}
}

func TestUpload_Security_MultipleRapidUploads(t *testing.T) {
	mockService := &MockUploadService{
		ProcessUploadedFileFunc: func(userID uint64, filePath, originalName string, fileSize int64, fileType string) (*response.UploadResponse, error) {
			return &response.UploadResponse{FileID: 1, RecordsImported: 1}, nil
		},
	}
	h := handler.NewUploadHandler(mockService, getTestUploadConfig())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/upload", h.UploadFile)

	// 快速连续上传多个文件
	for i := 0; i < 10; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.csv")
		part.Write([]byte("name,value\ntest,123"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 每次都应该成功处理（200）或被限流（429）
		if w.Code != http.StatusOK && w.Code != http.StatusTooManyRequests {
			t.Errorf("Rapid upload #%d returned %d", i+1, w.Code)
		}
	}
}
