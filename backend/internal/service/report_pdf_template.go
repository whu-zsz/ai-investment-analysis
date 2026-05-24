package service

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

	responsedto "stock-analysis-backend/internal/dto/response"
)

//go:embed templates/analysis_report_pdf.html
var analysisReportPDFTemplate string

type analysisReportPDFView struct {
	Title                   string
	ReportType              string
	AnalysisPeriod          string
	CreatedAt               string
	AIModel                 string
	TotalInvestment         string
	TotalProfit             string
	ProfitRate              string
	RiskLevel               string
	RiskLevelLabel          string
	InvestmentStyle         string
	MarketDataStatus        string
	SymbolsCount            int
	WinningTrades           int
	LosingTrades            int
	SummaryText             string
	RiskAnalysis            string
	PatternInsights         string
	PredictionText          string
	Recommendations         []string
	ProfitDistribution      []analysisReportPDFProfitPointView
	ProfitDistributionTop   string
	ProfitDistributionLow   string
	PositiveCount           int
	NegativeCount           int
	ZeroCount               int
	OutcomePositivePercent  string
	OutcomeNegativePercent  string
	OutcomeZeroPercent      string
	OutcomeGradient         string
	ProfitComposition       []analysisReportPDFCompositionView
	RealizedProfitTotal     string
	UnrealizedProfitTotal   string
	MomentumRanking         []analysisReportPDFMomentumView
	Items                   []analysisReportPDFItemView
}

type analysisReportPDFProfitPointView struct {
	Symbol        string
	Value         string
	SemanticLabel string
	BarClass      string
	BarWidth      string
}

type analysisReportPDFCompositionView struct {
	Symbol          string
	RealizedProfit  string
	UnrealizedProfit string
	TotalProfit     string
	BarWidth        string
	BarClass        string
}

type analysisReportPDFMomentumView struct {
	Symbol        string
	Value         string
	SemanticLabel string
	BarClass      string
	BarWidth      string
}

type analysisReportPDFItemView struct {
	Symbol            string
	AssetName         string
	Recommendation    string
	RecommendationTag string
	RiskLevel         string
	RiskLevelLabel    string
	InvestmentStyle   string
	TradeCount        int
	TotalProfit       string
	LatestPrice       string
	EndingPositionQty string
	AnalysisText      string
	KeyPoints         []string
}

func renderAnalysisReportPDFHTML(detail *responsedto.AnalysisReportDetailResponse) (string, error) {
	if detail == nil {
		return "", fmt.Errorf("analysis report detail is nil")
	}

	profitDistribution := buildProfitDistributionView(detail)
	positiveCount, negativeCount, zeroCount := summarizeProfitDistribution(profitDistribution)
	realizedTotal, unrealizedTotal := summarizeProfitComposition(detail.Items)
	compositionViews, maxCompositionAbs := buildProfitCompositionView(detail.Items)
	momentumRanking := buildMomentumRankingView(detail.Items)

	outcomePositivePercent, outcomeNegativePercent, outcomeZeroPercent, outcomeGradient := buildOutcomeStyle(positiveCount, negativeCount, zeroCount)

	view := analysisReportPDFView{
		Title:                  fallbackDisplay(detail.ReportTitle),
		ReportType:             fallbackDisplay(detail.ReportType),
		AnalysisPeriod:         fmt.Sprintf("%s ~ %s", fallbackDisplay(detail.AnalysisPeriodStart), fallbackDisplay(detail.AnalysisPeriodEnd)),
		CreatedAt:              fallbackDisplay(detail.CreatedAt),
		AIModel:                fallbackDisplay(detail.AIModel),
		TotalInvestment:        fallbackDisplay(detail.TotalInvestment),
		TotalProfit:            fallbackDisplay(detail.TotalProfit),
		ProfitRate:             fallbackDisplay(detail.ProfitRate),
		RiskLevel:              fallbackDisplay(detail.RiskLevel),
		RiskLevelLabel:         mapRiskLevelLabel(detail.RiskLevel),
		InvestmentStyle:        mapInvestmentStyleLabel(detail.InvestmentStyle),
		MarketDataStatus:       mapMarketStatusLabel(detail.MarketDataStatus),
		SymbolsCount:           detail.SymbolsCount,
		WinningTrades:          detail.WinningTrades,
		LosingTrades:           detail.LosingTrades,
		SummaryText:            fallbackParagraph(detail.SummaryText),
		RiskAnalysis:           fallbackParagraph(detail.RiskAnalysis),
		PatternInsights:        fallbackParagraph(detail.PatternInsights),
		PredictionText:         fallbackParagraph(detail.PredictionText),
		Recommendations:        fallbackList(detail.Recommendations),
		ProfitDistribution:     profitDistribution,
		ProfitDistributionTop:   summarizeExtremePoint(profitDistribution, true),
		ProfitDistributionLow:   summarizeExtremePoint(profitDistribution, false),
		PositiveCount:          positiveCount,
		NegativeCount:          negativeCount,
		ZeroCount:              zeroCount,
		OutcomePositivePercent: outcomePositivePercent,
		OutcomeNegativePercent: outcomeNegativePercent,
		OutcomeZeroPercent:     outcomeZeroPercent,
		OutcomeGradient:        outcomeGradient,
		ProfitComposition:      compositionViews,
		RealizedProfitTotal:    formatProfitFloat(realizedTotal),
		UnrealizedProfitTotal:  formatProfitFloat(unrealizedTotal),
		MomentumRanking:        momentumRanking,
		Items:                  make([]analysisReportPDFItemView, 0, len(detail.Items)),
	}

	_ = maxCompositionAbs

	for _, item := range detail.Items {
		view.Items = append(view.Items, analysisReportPDFItemView{
			Symbol:            fallbackDisplay(item.Symbol),
			AssetName:         fallbackDisplay(item.AssetName),
			Recommendation:    fallbackDisplay(item.Recommendation),
			RecommendationTag: strings.ToUpper(fallbackDisplay(item.Recommendation)),
			RiskLevel:         fallbackDisplay(item.RiskLevel),
			RiskLevelLabel:    mapRiskLevelLabel(item.RiskLevel),
			InvestmentStyle:   mapInvestmentStyleLabel(item.InvestmentStyle),
			TradeCount:        item.TradeCount,
			TotalProfit:       fallbackDisplay(item.TotalProfit),
			LatestPrice:       fallbackDisplay(item.LatestPrice),
			EndingPositionQty: fallbackDisplay(item.EndingPositionQty),
			AnalysisText:      fallbackParagraph(item.AnalysisText),
			KeyPoints:         fallbackList(item.KeyPoints),
		})
	}

	tmpl, err := template.New("analysis_report_pdf").Parse(analysisReportPDFTemplate)
	if err != nil {
		return "", fmt.Errorf("parse analysis report pdf template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, view); err != nil {
		return "", fmt.Errorf("execute analysis report pdf template: %w", err)
	}
	return buf.String(), nil
}

func buildAnalysisReportPDFFilename(detail *responsedto.AnalysisReportDetailResponse) string {
	if detail == nil {
		return "analysis-report.pdf"
	}
	return fmt.Sprintf("analysis-report-%d.pdf", detail.ID)
}

func buildProfitDistributionView(detail *responsedto.AnalysisReportDetailResponse) []analysisReportPDFProfitPointView {
	if detail == nil {
		return nil
	}

	points := parseProfitDistributionPoints(detail.ChartData)
	if len(points) == 0 {
		points = make([]chartPoint, 0, len(detail.Items))
		for _, item := range detail.Items {
			if strings.TrimSpace(item.Symbol) == "" {
				continue
			}
			value := strings.TrimSpace(item.RealizedProfit)
			if value == "" {
				value = strings.TrimSpace(item.TotalProfit)
			}
			if value == "" {
				continue
			}
			points = append(points, chartPoint{Symbol: item.Symbol, Value: value})
		}
	}
	if len(points) == 0 {
		return nil
	}

	type profitPoint struct {
		Symbol string
		Value  string
		Float  float64
	}

	normalized := make([]profitPoint, 0, len(points))
	maxAbs := 0.0
	for _, point := range points {
		value, err := strconv.ParseFloat(strings.TrimSpace(point.Value), 64)
		if err != nil {
			continue
		}
		absValue := value
		if absValue < 0 {
			absValue = -absValue
		}
		if absValue > maxAbs {
			maxAbs = absValue
		}
		normalized = append(normalized, profitPoint{
			Symbol: fallbackDisplay(point.Symbol),
			Value:  formatProfitFloat(value),
			Float:  value,
		})
	}
	if len(normalized) == 0 {
		return nil
	}
	if maxAbs == 0 {
		maxAbs = 1
	}

	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Float > normalized[j].Float
	})

	result := make([]analysisReportPDFProfitPointView, 0, len(normalized))
	for _, point := range normalized {
		semanticLabel := "持平"
		barClass := "bar-neutral"
		if point.Float > 0 {
			semanticLabel = "盈利"
			barClass = "bar-positive"
		} else if point.Float < 0 {
			semanticLabel = "亏损"
			barClass = "bar-negative"
		}

		width := (point.Float / maxAbs) * 100
		if width < 0 {
			width = -width
		}
		if width < 8 {
			width = 8
		}
		if width > 100 {
			width = 100
		}

		result = append(result, analysisReportPDFProfitPointView{
			Symbol:        point.Symbol,
			Value:         point.Value,
			SemanticLabel: semanticLabel,
			BarClass:      barClass,
			BarWidth:      fmt.Sprintf("%.0f%%", width),
		})
	}

	return result
}

func parseProfitDistributionPoints(chartData string) []chartPoint {
	trimmed := strings.TrimSpace(chartData)
	if trimmed == "" {
		return nil
	}

	var envelope chartDataEnvelope
	if err := json.Unmarshal([]byte(trimmed), &envelope); err == nil && envelope.Kind == "profit_by_symbol" {
		return envelope.Points
	}

	var legacyPoints []chartPoint
	if err := json.Unmarshal([]byte(trimmed), &legacyPoints); err == nil {
		return legacyPoints
	}

	return nil
}

func buildProfitCompositionView(items []responsedto.AnalysisReportItemResponse) ([]analysisReportPDFCompositionView, float64) {
	if len(items) == 0 {
		return nil, 0
	}

	type compositionPoint struct {
		symbol   string
		realized float64
		unreal   float64
		total    float64
	}

	points := make([]compositionPoint, 0, len(items))
	maxAbs := 0.0
	for _, item := range items {
		realized := toFloat64(item.RealizedProfit)
		unreal := toFloat64(item.UnrealizedProfit)
		total := realized + unreal
		if strings.TrimSpace(item.Symbol) == "" || (realized == 0 && unreal == 0) {
			continue
		}
		if absFloat(total) > maxAbs {
			maxAbs = absFloat(total)
		}
		points = append(points, compositionPoint{symbol: item.Symbol, realized: realized, unreal: unreal, total: total})
	}
	if len(points) == 0 {
		return nil, 0
	}
	if maxAbs == 0 {
		maxAbs = 1
	}

	sort.Slice(points, func(i, j int) bool {
		return absFloat(points[i].total) > absFloat(points[j].total)
	})
	if len(points) > 6 {
		points = points[:6]
	}

	result := make([]analysisReportPDFCompositionView, 0, len(points))
	for _, point := range points {
		width := absFloat(point.total) / maxAbs * 100
		if width < 8 {
			width = 8
		}
		if width > 100 {
			width = 100
		}
		barClass := "bar-neutral"
		if point.total > 0 {
			barClass = "bar-positive"
		} else if point.total < 0 {
			barClass = "bar-negative"
		}
		result = append(result, analysisReportPDFCompositionView{
			Symbol:           fallbackDisplay(point.symbol),
			RealizedProfit:   formatProfitFloat(point.realized),
			UnrealizedProfit: formatProfitFloat(point.unreal),
			TotalProfit:      formatProfitFloat(point.total),
			BarWidth:         fmt.Sprintf("%.0f%%", width),
			BarClass:         barClass,
		})
	}

	return result, maxAbs
}

func summarizeProfitComposition(items []responsedto.AnalysisReportItemResponse) (float64, float64) {
	realizedTotal := 0.0
	unrealizedTotal := 0.0
	for _, item := range items {
		realizedTotal += toFloat64(item.RealizedProfit)
		unrealizedTotal += toFloat64(item.UnrealizedProfit)
	}
	return realizedTotal, unrealizedTotal
}

func buildMomentumRankingView(items []responsedto.AnalysisReportItemResponse) []analysisReportPDFMomentumView {
	if len(items) == 0 {
		return nil
	}

	type momentumPoint struct {
		symbol string
		value  float64
	}

	points := make([]momentumPoint, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Symbol) == "" || strings.TrimSpace(item.ChangePercent7D) == "" {
			continue
		}
		points = append(points, momentumPoint{symbol: item.Symbol, value: toFloat64(item.ChangePercent7D)})
	}
	if len(points) == 0 {
		return nil
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].value > points[j].value
	})
	if len(points) > 6 {
		points = points[:6]
	}

	maxAbs := absFloat(points[0].value)
	for _, point := range points[1:] {
		if absFloat(point.value) > maxAbs {
			maxAbs = absFloat(point.value)
		}
	}
	if maxAbs == 0 {
		maxAbs = 1
	}

	result := make([]analysisReportPDFMomentumView, 0, len(points))
	for _, point := range points {
		semanticLabel := "持平"
		barClass := "bar-neutral"
		if point.value > 0 {
			semanticLabel = "上升"
			barClass = "bar-positive"
		} else if point.value < 0 {
			semanticLabel = "下跌"
			barClass = "bar-negative"
		}
		width := absFloat(point.value) / maxAbs * 100
		if width < 8 {
			width = 8
		}
		if width > 100 {
			width = 100
		}
		result = append(result, analysisReportPDFMomentumView{
			Symbol:        fallbackDisplay(point.symbol),
			Value:         formatProfitFloat(point.value) + "%",
			SemanticLabel: semanticLabel,
			BarClass:      barClass,
			BarWidth:      fmt.Sprintf("%.0f%%", width),
		})
	}
	return result
}

func buildOutcomeStyle(positiveCount, negativeCount, zeroCount int) (string, string, string, string) {
	total := positiveCount + negativeCount + zeroCount
	if total == 0 {
		return "0%", "0%", "0%", "conic-gradient(#d9d9d9 0 100%)"
	}
	positivePercent := positiveCount * 100 / total
	negativePercent := negativeCount * 100 / total
	zeroPercent := 100 - positivePercent - negativePercent
	return fmt.Sprintf("%d%%", positivePercent), fmt.Sprintf("%d%%", negativePercent), fmt.Sprintf("%d%%", zeroPercent), fmt.Sprintf("conic-gradient(#52c41a 0 %d%%, #ff4d4f %d%% %d%%, #bfbfbf %d%% 100%%)", positivePercent, positivePercent, positivePercent+negativePercent, positivePercent+negativePercent)
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func toFloat64(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
}

func summarizeProfitDistribution(points []analysisReportPDFProfitPointView) (int, int, int) {
	positiveCount := 0
	negativeCount := 0
	zeroCount := 0
	for _, point := range points {
		switch point.SemanticLabel {
		case "盈利":
			positiveCount++
		case "亏损":
			negativeCount++
		default:
			zeroCount++
		}
	}
	return positiveCount, negativeCount, zeroCount
}

func summarizeExtremePoint(points []analysisReportPDFProfitPointView, top bool) string {
	if len(points) == 0 {
		return "—"
	}
	if top {
		return fmt.Sprintf("%s %s", points[0].Symbol, points[0].Value)
	}
	last := points[len(points)-1]
	return fmt.Sprintf("%s %s", last.Symbol, last.Value)
}

func formatProfitFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func fallbackDisplay(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func fallbackParagraph(value string) string {
	if strings.TrimSpace(value) == "" {
		return "暂无内容。"
	}
	return value
}

func fallbackList(values []string) []string {
	if len(values) == 0 {
		return []string{"暂无内容"}
	}
	return values
}

func mapRiskLevelLabel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "low":
		return "低风险"
	case "high":
		return "高风险"
	default:
		return "中风险"
	}
}

func mapInvestmentStyleLabel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "aggressive":
		return "激进型成长"
	case "conservative":
		return "保守防御"
	case "balanced":
		return "稳健均衡"
	default:
		return fallbackDisplay(value)
	}
}

func mapMarketStatusLabel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case marketDataStatusComplete:
		return "市场数据完整"
	case marketDataStatusFetchedLive:
		return "实时补充成功"
	case marketDataStatusPartial:
		return "市场数据部分缺失"
	case marketDataStatusUnavailable:
		return "市场数据不可用"
	default:
		return fallbackDisplay(value)
	}
}
