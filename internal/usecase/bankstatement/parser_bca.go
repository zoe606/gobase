package bankstatement

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BCAParser struct{}

func (p *BCAParser) Parse(text string) (*ParseResult, error) {
	lines := strings.Split(text, "\n")
	var items []ParsedItem
	var periodStart, periodEnd *time.Time

	dateRegex := regexp.MustCompile(`^(\d{2}/\d{2})\s+(.+?)\s+([\d,]+\.\d{2})\s+(DB|CR)\s+([\d,]+\.\d{2})`)

	currentYear := time.Now().Year()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := dateRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		dateStr := matches[1]
		description := strings.TrimSpace(matches[2])
		amountStr := strings.ReplaceAll(matches[3], ",", "")
		txType := matches[4]
		balanceStr := strings.ReplaceAll(matches[5], ",", "")

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			continue
		}

		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			continue
		}

		parts := strings.Split(dateStr, "/")
		if len(parts) != 2 {
			continue
		}
		month, _ := strconv.Atoi(parts[1])
		day, _ := strconv.Atoi(parts[0])
		txDate := time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local)

		item := ParsedItem{
			Date:        txDate,
			Description: description,
			Balance:     balance,
		}

		if txType == "DB" {
			item.Debit = amount
		} else {
			item.Credit = amount
		}

		items = append(items, item)

		if periodStart == nil || txDate.Before(*periodStart) {
			t := txDate
			periodStart = &t
		}
		if periodEnd == nil || txDate.After(*periodEnd) {
			t := txDate
			periodEnd = &t
		}
	}

	return &ParseResult{
		Items:       items,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}, nil
}
