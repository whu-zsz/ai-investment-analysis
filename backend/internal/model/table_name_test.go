package model

import (
	"testing"
)

func TestUser_TableName(t *testing.T) {
	u := User{}
	if u.TableName() != "users" {
		t.Errorf("User.TableName() = %v, want %v", u.TableName(), "users")
	}
}

func TestTransaction_TableName(t *testing.T) {
	tx := Transaction{}
	if tx.TableName() != "transactions" {
		t.Errorf("Transaction.TableName() = %v, want %v", tx.TableName(), "transactions")
	}
}

func TestMarketSnapshot_TableName(t *testing.T) {
	ms := MarketSnapshot{}
	if ms.TableName() != "market_snapshots" {
		t.Errorf("MarketSnapshot.TableName() = %v, want %v", ms.TableName(), "market_snapshots")
	}
}

func TestAnalysisTask_TableName(t *testing.T) {
	at := AnalysisTask{}
	if at.TableName() != "ai_analysis_tasks" {
		t.Errorf("AnalysisTask.TableName() = %v, want %v", at.TableName(), "ai_analysis_tasks")
	}
}

func TestAnalysisReport_TableName(t *testing.T) {
	ar := AnalysisReport{}
	if ar.TableName() != "ai_analysis_reports" {
		t.Errorf("AnalysisReport.TableName() = %v, want %v", ar.TableName(), "ai_analysis_reports")
	}
}

func TestAnalysisReportItem_TableName(t *testing.T) {
	ari := AnalysisReportItem{}
	if ari.TableName() != "ai_analysis_report_items" {
		t.Errorf("AnalysisReportItem.TableName() = %v, want %v", ari.TableName(), "ai_analysis_report_items")
	}
}

func TestStockAnalysisMetric_TableName(t *testing.T) {
	sam := StockAnalysisMetric{}
	if sam.TableName() != "stock_analysis_metrics" {
		t.Errorf("StockAnalysisMetric.TableName() = %v, want %v", sam.TableName(), "stock_analysis_metrics")
	}
}

func TestUploadedFile_TableName(t *testing.T) {
	uf := UploadedFile{}
	if uf.TableName() != "uploaded_files" {
		t.Errorf("UploadedFile.TableName() = %v, want %v", uf.TableName(), "uploaded_files")
	}
}
