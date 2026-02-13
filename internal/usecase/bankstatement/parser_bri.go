package bankstatement

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BRIParser struct{}

func (p *BRIParser) Parse(text string) (*ParseResult, error) {
	lines := strings.Split(text, "\n")
	var items []ParsedItem
	var periodStart, periodEnd *time.Time

	dateRegex := regexp.MustCompile(`^(\d{2}[/-]\d{2}[/-]\d{4})\s+(.+?)\s+([\d.,]+)\s+([\d.,]+)\s+([\d.,]+)`)

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
		debitStr := strings.ReplaceAll(strings.ReplaceAll(matches[3], ".", ""), ",", ".")
		creditStr := strings.ReplaceAll(strings.ReplaceAll(matches[4], ".", ""), ",", ".")
		balanceStr := strings.ReplaceAll(strings.ReplaceAll(matches[5], ".", ""), ",", ".")

		debit, _ := strconv.ParseFloat(debitStr, 64)
		credit, _ := strconv.ParseFloat(creditStr, 64)
		balance, _ := strconv.ParseFloat(balanceStr, 64)

		dateStr = strings.ReplaceAll(dateStr, "-", "/")
		txDate, err := time.Parse("02/01/2006", dateStr)
		if err != nil {
			continue
		}

		items = append(items, ParsedItem{
			Date:        txDate,
			Description: description,
			Debit:       debit,
			Credit:      credit,
			Balance:     balance,
		})

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
