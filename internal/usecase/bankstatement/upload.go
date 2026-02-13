package bankstatement

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	"go-boilerplate/internal/entity"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

var installmentKeywords = []string{"CICILAN", "INSTALLMENT", "ANGSURAN", "ANGSRN"}

func (uc *UseCase) Upload(ctx context.Context, req bankstatementdto.UploadRequest) (*bankstatementdto.ResponseWithItems, error) {
	bank, err := uc.bankRepo.GetByID(ctx, req.BankID)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - Upload - bankRepo.GetByID: %w", err)
	}

	parser, ok := uc.parsers[bank.Code]
	if !ok {
		return nil, ErrUnsupportedBank
	}

	password := resolvePassword(req.Password, bank.DefaultPassword)

	content, err := io.ReadAll(req.File)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - Upload - ReadAll: %w", err)
	}

	text, err := extractTextFromPDF(content, password)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - Upload - extractText: %w", err)
	}

	parseResult, err := parser.Parse(text)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - Upload - parser.Parse: %w", err)
	}

	stmt := buildStatement(req, password, parseResult)

	if err := uc.stmtRepo.Create(ctx, stmt); err != nil {
		return nil, fmt.Errorf("bankstatement - Upload - stmtRepo.Create: %w", err)
	}

	lineItems, totalDebit, totalCredit := buildLineItems(stmt.ID, parseResult.Items)

	if len(lineItems) > 0 {
		if err := uc.lineItemRepo.BulkCreate(ctx, lineItems); err != nil {
			return nil, fmt.Errorf("bankstatement - Upload - lineItemRepo.BulkCreate: %w", err)
		}
	}

	stmt.TotalDebit = totalDebit
	stmt.TotalCredit = totalCredit
	stmt.Bank = bank
	if err := uc.stmtRepo.Update(ctx, stmt); err != nil {
		return nil, fmt.Errorf("bankstatement - Upload - stmtRepo.Update: %w", err)
	}

	return bankstatementdto.NewResponseWithItems(stmt, lineItems), nil
}

func resolvePassword(override string, defaultPW *string) string {
	if override != "" {
		return override
	}
	if defaultPW != nil {
		return *defaultPW
	}
	return ""
}

func buildStatement(req bankstatementdto.UploadRequest, password string, result *ParseResult) *entity.BankStatement {
	stmt := &entity.BankStatement{
		UserID: req.UserID,
		BankID: req.BankID,
		Status: entity.BankStatementStatusCompleted,
	}
	if password != "" {
		stmt.Password = &password
	}
	if result.PeriodStart != nil {
		s := result.PeriodStart.Format("2006-01-02")
		stmt.PeriodStart = &s
	}
	if result.PeriodEnd != nil {
		s := result.PeriodEnd.Format("2006-01-02")
		stmt.PeriodEnd = &s
	}
	return stmt
}

func buildLineItems(stmtID uint, items []ParsedItem) (lineItems []*entity.LineItem, totalDebit, totalCredit float64) {
	lineItems = make([]*entity.LineItem, 0, len(items))
	for _, item := range items {
		totalDebit += item.Debit
		totalCredit += item.Credit

		li := &entity.LineItem{
			SourceType:    "bank_statement",
			SourceID:      stmtID,
			Date:          item.Date.Format("2006-01-02"),
			Description:   item.Description,
			Debit:         item.Debit,
			Credit:        item.Credit,
			Balance:       item.Balance,
			IsInstallment: isInstallmentTransaction(item.Description),
		}
		lineItems = append(lineItems, li)
	}
	return lineItems, totalDebit, totalCredit
}

func extractTextFromPDF(content []byte, password string) (string, error) {
	reader := bytes.NewReader(content)

	conf := model.NewDefaultConfiguration()
	if password != "" {
		conf.UserPW = password
		conf.OwnerPW = password
	}

	ctx, err := api.ReadValidateAndOptimize(reader, conf)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidPDF, err)
	}

	var allText strings.Builder
	for i := 1; i <= ctx.PageCount; i++ {
		r, err := api.ExtractPage(ctx, i)
		if err != nil {
			continue
		}
		pageBytes, err := io.ReadAll(r)
		if err != nil {
			continue
		}
		allText.Write(pageBytes)
		allText.WriteString("\n")
	}

	return allText.String(), nil
}

func isInstallmentTransaction(description string) bool {
	upper := strings.ToUpper(description)
	for _, keyword := range installmentKeywords {
		if strings.Contains(upper, keyword) {
			return true
		}
	}
	return false
}
