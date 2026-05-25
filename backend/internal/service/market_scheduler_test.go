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
	BatchNo          string
	Count            int
	SnapshotErr      error
	FullSnapshotErr  error
	BoardSnapshotErr error
	TrackedCalled    bool
	FullMarketCalled bool
	BoardCalled      bool
	WarmupCalled     bool
}

func (s *MockSchedulerMarketDataService) FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error) {
	s.TrackedCalled = true
	return s.BatchNo, s.Count, s.SnapshotErr
}

func (s *MockSchedulerMarketDataService) FetchAndStoreFullMarketSnapshots(ctx context.Context) (string, int, error) {
	s.FullMarketCalled = true
	return s.BatchNo, s.Count, s.FullSnapshotErr
}

func (s *MockSchedulerMarketDataService) FetchAndStoreMarketBoardSnapshots(ctx context.Context) (string, int, error) {
	s.BoardCalled = true
	return s.BatchNo, s.Count, s.BoardSnapshotErr
}

func (s *MockSchedulerMarketDataService) FetchAndStoreQuotesBySymbols(ctx context.Context, symbols []string) ([]model.MarketSnapshot, error) {
	return nil, nil
}

func (s *MockSchedulerMarketDataService) EnsureTrackedIndexHistory(ctx context.Context) error {
	s.WarmupCalled = true
	return s.SnapshotErr
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

	scheduler.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()

	if !mockSvc.WarmupCalled {
		t.Error("MarketScheduler should have called EnsureTrackedIndexHistory")
	}
	if !mockSvc.TrackedCalled {
		t.Error("MarketScheduler should have called FetchAndStoreMarketSnapshots")
	}
	if !mockSvc.FullMarketCalled {
		t.Error("MarketScheduler should have called FetchAndStoreFullMarketSnapshots")
	}
	if !mockSvc.BoardCalled {
		t.Error("MarketScheduler should have called FetchAndStoreMarketBoardSnapshots")
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

	if !mockSvc.TrackedCalled {
		t.Error("runOnce() should call FetchAndStoreMarketSnapshots")
	}
	if !mockSvc.FullMarketCalled {
		t.Error("runOnce() should call FetchAndStoreFullMarketSnapshots")
	}
	if !mockSvc.BoardCalled {
		t.Error("runOnce() should call FetchAndStoreMarketBoardSnapshots")
	}
}

// TestMarketScheduler_RunOnce_Error 测试单次运行错误
func TestMarketScheduler_RunOnce_Error(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{
		SnapshotErr:      context.Canceled,
		FullSnapshotErr:  context.Canceled,
		BoardSnapshotErr: context.Canceled,
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

	if !mockSvc.TrackedCalled {
		t.Error("runOnce() should call FetchAndStoreMarketSnapshots even on error")
	}
	if !mockSvc.FullMarketCalled {
		t.Error("runOnce() should call FetchAndStoreFullMarketSnapshots even on error")
	}
	if !mockSvc.BoardCalled {
		t.Error("runOnce() should call FetchAndStoreMarketBoardSnapshots even on error")
	}
}

// TestMarketScheduler_Interface 测试接口实现
func TestMarketScheduler_Interface(t *testing.T) {
	mockSvc := &MockSchedulerMarketDataService{}
	logger := zap.NewNop()

	var _ MarketScheduler = NewMarketScheduler(time.Minute, mockSvc, logger)
}
