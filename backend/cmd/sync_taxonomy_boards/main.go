package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type stockRow struct {
	Symbol         string
	Name           string
	Industry       string
	Concepts       string
	TotalMarketCap decimal.Decimal
	FloatMarketCap decimal.Decimal
}

func main() {
	var (
		sourceTag = flag.String("source", "taxonomy-sync", "source tag written to market_board_constituents")
	)
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("init db failed: %v", err)
	}
	defer func() {
		_ = config.CloseDB(db)
	}()

	rows, err := loadTaxonomyRows(db)
	if err != nil {
		log.Fatalf("load taxonomy rows failed: %v", err)
	}
	if len(rows) == 0 {
		log.Println("no taxonomy rows available")
		return
	}

	existing, err := loadExistingConstituents(db)
	if err != nil {
		log.Fatalf("load existing constituents failed: %v", err)
	}

	now := time.Now()
	upserts := make([]model.MarketBoardConstituent, 0, len(rows)*2)
	for _, row := range rows {
		symbol := normalizeSymbol(row.Symbol)
		if symbol == "" || !isStockSymbol(symbol) {
			continue
		}

		industry := model.NormalizeIndustryLabel(row.Industry)
		if industry != "" {
			key := buildKey("industry", industry, symbol)
			item, ok := existing[key]
			if !ok {
				item = model.MarketBoardConstituent{
					BoardType: "industry",
					BoardCode: buildBoardCode("industry", industry),
					BoardName: industry,
					Symbol:    symbol,
				}
			}
			item.StockName = fallbackText(item.StockName, strings.TrimSpace(row.Name))
			item.TotalMarketCap = maxDecimal(item.TotalMarketCap, row.TotalMarketCap)
			item.FloatMarketCap = maxDecimal(item.FloatMarketCap, row.FloatMarketCap)
			item.Source = mergeSource(item.Source, *sourceTag)
			item.SyncedAt = now
			upserts = append(upserts, item)
		}

		for _, concept := range model.NormalizeConceptList(strings.Split(row.Concepts, ",")) {
			key := buildKey("concept", concept, symbol)
			item, ok := existing[key]
			if !ok {
				item = model.MarketBoardConstituent{
					BoardType: "concept",
					BoardCode: buildBoardCode("concept", concept),
					BoardName: concept,
					Symbol:    symbol,
				}
			}
			item.StockName = fallbackText(item.StockName, strings.TrimSpace(row.Name))
			item.TotalMarketCap = maxDecimal(item.TotalMarketCap, row.TotalMarketCap)
			item.FloatMarketCap = maxDecimal(item.FloatMarketCap, row.FloatMarketCap)
			item.Source = mergeSource(item.Source, *sourceTag)
			item.SyncedAt = now
			upserts = append(upserts, item)
		}
	}

	if len(upserts) == 0 {
		log.Println("nothing to upsert")
		return
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(upserts); start += 500 {
			end := start + 500
			if end > len(upserts) {
				end = len(upserts)
			}
			for _, item := range upserts[start:end] {
				var existingRow model.MarketBoardConstituent
				err := tx.Where("board_type = ? AND board_code = ? AND symbol = ?", item.BoardType, item.BoardCode, item.Symbol).First(&existingRow).Error
				if err == nil {
					existingRow.BoardName = item.BoardName
					existingRow.StockName = item.StockName
					existingRow.TotalMarketCap = item.TotalMarketCap
					existingRow.FloatMarketCap = item.FloatMarketCap
					existingRow.Source = item.Source
					existingRow.SyncedAt = item.SyncedAt
					if saveErr := tx.Save(&existingRow).Error; saveErr != nil {
						return saveErr
					}
					continue
				}
				if err != nil && err != gorm.ErrRecordNotFound {
					return err
				}
				if createErr := tx.Create(&item).Error; createErr != nil {
					return createErr
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("sync taxonomy boards failed: %v", err)
	}

	log.Printf("sync taxonomy boards done: upserts=%d", len(upserts))
}

func loadTaxonomyRows(db *gorm.DB) ([]stockRow, error) {
	var rows []stockRow
	err := db.Model(&model.StockQuoteDetail{}).
		Select("symbol", "name", "industry", "concepts", "total_market_cap", "float_market_cap").
		Where("TRIM(IFNULL(industry,'')) <> '' OR TRIM(IFNULL(concepts,'')) <> ''").
		Find(&rows).Error
	return rows, err
}

func loadExistingConstituents(db *gorm.DB) (map[string]model.MarketBoardConstituent, error) {
	var rows []model.MarketBoardConstituent
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]model.MarketBoardConstituent, len(rows))
	for _, row := range rows {
		result[buildKey(row.BoardType, row.BoardName, row.Symbol)] = row
	}
	return result, nil
}

func buildKey(boardType, boardName, symbol string) string {
	return strings.ToLower(strings.TrimSpace(boardType)) + "|" + strings.TrimSpace(boardName) + "|" + normalizeSymbol(symbol)
}

func buildBoardCode(boardType, boardName string) string {
	slug := strings.ToLower(strings.TrimSpace(boardName))
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", "-", "_", ".", "_", "(", "", ")", "", "（", "", "）", "", "、", "_", ",", "_")
	slug = replacer.Replace(slug)
	if slug == "" {
		slug = "unknown"
	}
	return boardType + "_tax_" + slug
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func isStockSymbol(symbol string) bool {
	return strings.HasSuffix(symbol, ".SH") || strings.HasSuffix(symbol, ".SZ") || strings.HasSuffix(symbol, ".BJ")
}

func fallbackText(current, next string) string {
	if strings.TrimSpace(current) != "" {
		return current
	}
	return next
}

func maxDecimal(a, b decimal.Decimal) decimal.Decimal {
	if b.GreaterThan(a) {
		return b
	}
	return a
}

func mergeSource(existing, extra string) string {
	existing = strings.TrimSpace(existing)
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return existing
	}
	if existing == "" {
		return extra
	}
	for _, part := range strings.Split(existing, "+") {
		if strings.TrimSpace(part) == extra {
			return existing
		}
	}
	merged := existing + "+" + extra
	if len(merged) > 32 {
		return existing
	}
	return merged
}
