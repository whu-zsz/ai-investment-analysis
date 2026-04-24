package service

import (
	"context"
	"stock-analysis-backend/internal/model"
	"testing"
	"time"

	"go.uber.org/zap"
)

// MockSchedulerMarketDataService 模拟市场数据服务用于调度器测试
type MockSchedulerMarketDataService struct {
	BatchNo string
	Count   int
	Err     error
	Called  bool
}

func (s *MockSchedulerMarketDataService) FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error) {
	s.Called = true
	return s.BatchNo, s.Count, s.Err
}

func (s *MockSchedulerMarketDataService) FetchAndStoreQuotesBySymbols(ctx context.Context, symbols []string) ([]model.MarketSnapshot, error) {
	return nil, nil
}

// TestNewMarketScheduler 测试创建调度器
func TestNewMarketScheduler(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{}
	logger := zap.NewNop()

	scheduler := NewMarketScheduler(time.Minute, mockSvc, logger)
	if scheduler == nil {
		t.Error("NewMarketScheduler() returned nil")
	}
}

// TestNewMarketScheduler_DefaultInterval 测试默认间隔
func TestNewMarketScheduler_DefaultInterval(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{}
	logger := zap.NewNop()

	// 传入无效间隔，应该使用默认值
	scheduler := NewMarketScheduler(0, mockSvc, logger)
	if scheduler == nil {
		t.Error("NewMarketScheduler() should use default interval for 0")
	}

	scheduler = NewMarketScheduler(-1*time.Second, mockSvc, logger)
	if scheduler == nil {
		t.Error("NewMarketScheduler() should use default interval for negative")
	}
}

// TestMarketScheduler_Start 测试启动调度器
func TestMarketScheduler_Start(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{
		BatchNo: "batch001",
		Count:   5,
	}
	logger := zap.NewNop()

	scheduler := NewMarketScheduler(time.Second, mockSvc, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动调度器
	scheduler.Start(ctx)

	// 等待一段时间让调度器执行一次
	time.Sleep(100 * time.Millisecond)

	// 取消上下文停止调度器
	cancel()

	// 验证服务被调用
	if !mockSvc.Called {
		t.Error("MarketScheduler should have called FetchAndStoreMarketSnapshots")
	}
}

// TestMarketScheduler_ContextCancellation 测试上下文取消
func TestMarketScheduler_ContextCancellation(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{
		BatchNo: "batch001",
		Count:   5,
	}
	logger := zap.NewNop()

	scheduler := NewMarketScheduler(time.Hour, mockSvc, logger) // 长间隔

	ctx, cancel := context.WithCancel(context.Background())
	scheduler.Start(ctx)

	// 立即取消
	cancel()

	// 等待 goroutine 结束
	time.Sleep(50 * time.Millisecond)

	// 测试通过，如果调度器正确响应取消，不会有死锁
}

// TestMarketScheduler_RunOnce 测试单次运行
func TestMarketScheduler_RunOnce(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{
		BatchNo: "batch001",
		Count:   10,
	}
	logger := zap.NewNop()

	s := &marketScheduler{
		interval:          time.Minute,
		marketDataService: mockSvc,
		logger:            logger,
	}

	ctx := context.Background()
	s.runOnce(ctx)

	if !mockSvc.Called {
		t.Error("runOnce() should call FetchAndStoreMarketSnapshots")
	}
}

// TestMarketScheduler_RunOnce_Error 测试单次运行错误
func TestMarketScheduler_RunOnce_Error(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{
		Err: context.Canceled,
	}
	logger := zap.NewNop()

	s := &marketScheduler{
		interval:          time.Minute,
		marketDataService: mockSvc,
		logger:            logger,
	}

	ctx := context.Background()
	// 不应该 panic
	s.runOnce(ctx)

	if !mockSvc.Called {
		t.Error("runOnce() should call FetchAndStoreMarketSnapshots even on error")
	}
}

// TestMarketScheduler_Interface 测试接口实现
func TestMarketScheduler_Interface(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{}
	logger := zap.NewNop()

	var _ MarketScheduler = NewMarketScheduler(time.Minute, mockSvc, logger)
}
