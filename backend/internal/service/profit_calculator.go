package service

import (
	"sort"
	"strings"

	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"

	"github.com/shopspring/decimal"
)

type realizedProfitSummary struct {
	TotalTransactions int64
	BuyCount          int64
	SellCount         int64
	TotalInvestment   decimal.Decimal
	TotalProfit       decimal.Decimal
}

type inventoryLot struct {
	Quantity decimal.Decimal
	Cost     decimal.Decimal
}

type assetState struct {
	lots []inventoryLot
}

func buildRealizedProfitSnapshots(transactions []model.Transaction) map[uint64]decimal.Decimal {
	ordered := make([]model.Transaction, 0, len(transactions))
	ordered = append(ordered, filterStockTransactions(transactions)...)
	sortTransactionsChronologically(ordered)

	states := make(map[string]*assetState)
	profits := make(map[uint64]decimal.Decimal, len(ordered))

	for _, tx := range ordered {
		symbol := normalizeSymbol(tx.AssetCode)
		if symbol == "" {
			continue
		}
		state := states[symbol]
		if state == nil {
			state = &assetState{}
			states[symbol] = state
		}

		switch strings.ToLower(strings.TrimSpace(tx.TransactionType)) {
		case "buy":
			state.lots = append(state.lots, inventoryLot{
				Quantity: tx.Quantity,
				Cost:     tx.TotalAmount,
			})
		case "sell":
			profits[tx.ID] = calculateSellRealizedProfit(state, tx)
		case "dividend":
			continue
		}
	}

	return profits
}

func buildRealizedProfitSummary(transactions []model.Transaction) realizedProfitSummary {
	ordered := filterStockTransactions(transactions)
	snapshots := buildRealizedProfitSnapshots(ordered)
	summary := realizedProfitSummary{TotalInvestment: decimal.Zero, TotalProfit: decimal.Zero}

	for _, tx := range ordered {
		summary.TotalTransactions++
		switch strings.ToLower(strings.TrimSpace(tx.TransactionType)) {
		case "buy":
			summary.BuyCount++
			summary.TotalInvestment = summary.TotalInvestment.Add(tx.TotalAmount)
		case "sell":
			summary.SellCount++
			if profit, ok := snapshots[tx.ID]; ok {
				summary.TotalProfit = summary.TotalProfit.Add(profit)
			}
		}
	}

	return summary
}

func buildTransactionResponsesWithRealizedProfit(transactions []model.Transaction) []response.TransactionResponse {
	snapshots := buildRealizedProfitSnapshots(transactions)
	return buildTransactionResponsesWithProfitSnapshots(transactions, snapshots)
}

func buildTransactionResponsesWithProfitSnapshots(transactions []model.Transaction, snapshots map[uint64]decimal.Decimal) []response.TransactionResponse {
	result := make([]response.TransactionResponse, 0, len(transactions))
	for i := range transactions {
		resp := response.NewTransactionResponse(&transactions[i])
		if profit, ok := snapshots[transactions[i].ID]; ok {
			profitStr := profit.StringFixed(2)
			resp.Profit = &profitStr
		}
		result = append(result, resp)
	}
	return result
}

func sortTransactionsChronologically(transactions []model.Transaction) {
	sort.SliceStable(transactions, func(i, j int) bool {
		if transactions[i].TransactionDate.Equal(transactions[j].TransactionDate) {
			if transactions[i].CreatedAt.Equal(transactions[j].CreatedAt) {
				return transactions[i].ID < transactions[j].ID
			}
			return transactions[i].CreatedAt.Before(transactions[j].CreatedAt)
		}
		return transactions[i].TransactionDate.Before(transactions[j].TransactionDate)
	})
}

func calculateSellRealizedProfit(state *assetState, tx model.Transaction) decimal.Decimal {
	remaining := tx.Quantity
	costBasis := decimal.Zero
	lots := make([]inventoryLot, 0, len(state.lots))

	for _, lot := range state.lots {
		if remaining.LessThanOrEqual(decimal.Zero) {
			lots = append(lots, lot)
			continue
		}
		if lot.Quantity.LessThanOrEqual(decimal.Zero) {
			continue
		}

		consume := decimal.Min(remaining, lot.Quantity)
		if !lot.Quantity.IsZero() {
			consumedCost := lot.Cost.Mul(consume).Div(lot.Quantity)
			costBasis = costBasis.Add(consumedCost)
			lot.Cost = lot.Cost.Sub(consumedCost)
		}
		lot.Quantity = lot.Quantity.Sub(consume)
		remaining = remaining.Sub(consume)
		if lot.Quantity.GreaterThan(decimal.Zero) {
			lots = append(lots, lot)
		}
	}
	state.lots = lots

	if remaining.GreaterThan(decimal.Zero) {
		return decimal.Zero
	}
	return tx.TotalAmount.Sub(costBasis)
}
