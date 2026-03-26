package handler

import (
	"os"
	"path/filepath"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	uploadService service.UploadService
	uploadCfg     config.UploadConfig
}

func NewUploadHandler(uploadService service.UploadService, uploadCfg config.UploadConfig) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
		uploadCfg:     uploadCfg,
	}
}

// UploadFile godoc
// @Summary 上传投资记录文件
// @Description 上传CSV或Excel格式的投资记录文件
// @Tags 文件上传
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "CSV或Excel文件"
// @Success 200 {object} response.Response{data=response.UploadResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/upload [post]
func (h *UploadHandler) UploadFile(c *gin.Context) {
	userID := c.GetUint64("user_id")

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	// 验证文件类型
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	if fileExt != ".csv" && fileExt != ".xlsx" && fileExt != ".xls" {
		response.BadRequest(c, "unsupported file type, only CSV and Excel files are allowed")
		return
	}

	// 验证文件大小
	if file.Size > h.uploadCfg.MaxUploadSize {
		response.BadRequest(c, "file size exceeds maximum limit")
		return
	}

	// 确保上传目录存在
	if err := os.MkdirAll(h.uploadCfg.Path, 0755); err != nil {
		response.InternalServerError(c, "failed to create upload directory")
		return
	}

	// 生成唯一文件名
	fileUUID := uuid.New().String()
	newFileName := fileUUID + fileExt
	filePath := filepath.Join(h.uploadCfg.Path, newFileName)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		response.InternalServerError(c, "failed to save file")
		return
	}

	// 处理上传文件
	fileType := strings.TrimPrefix(fileExt, ".")
	uploadResp, err := h.uploadService.ProcessUploadedFile(userID, filePath, file.Filename, file.Size, fileType)
	if err != nil {
		os.Remove(filePath) // 删除已保存的文件
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, uploadResp)
}

// GetUploadHistory godoc
// @Summary 获取上传历史
// @Description 获取用户的文件上传历史记录
// @Tags 文件上传
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]response.UploadHistoryResponse}
// @Router /api/v1/upload/history [get]
func (h *UploadHandler) GetUploadHistory(c *gin.Context) {
	userID := c.GetUint64("user_id")

	history, err := h.uploadService.GetUploadHistory(userID)
	if err != nil {
		response.InternalServerError(c, "failed to get upload history")
		return
	}

	response.Success(c, history)
}
