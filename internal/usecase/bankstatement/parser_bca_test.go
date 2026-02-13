package bankstatement

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBCAParser_Parse(t *testing.T) {
	t.Parallel()

	currentYear := time.Now().Year()

	tests := []struct {
		name            string
		text            string
		wantItemCount   int
		wantPeriodStart *time.Time
		wantPeriodEnd   *time.Time
		wantItems       []ParsedItem
		wantErr         bool
	}{
		{
			name: "valid BCA statement with multiple transactions",
			text: `REKENING : 1234567890
PT COMPANY NAME
PERIODE : 01/01/2025 - 31/01/2025

TANGGAL    KETERANGAN                      MUTASI        SALDO
01/01  TRSF E-BANKING DB 01/01           500,000.00 DB  10,000,000.00
05/01  TRSF E-BANKING CR 05/01         1,200,000.00 CR  11,200,000.00
15/01  PAYMENT ELECTRICITY              350,000.00 DB  10,850,000.00
25/01  SALARY JANUARY 2025           15,000,000.00 CR  25,850,000.00
`,
			wantItemCount: 4,
			wantPeriodStart: func() *time.Time {
				t := time.Date(currentYear, time.January, 1, 0, 0, 0, 0, time.Local)
				return &t
			}(),
			wantPeriodEnd: func() *time.Time {
				t := time.Date(currentYear, time.January, 25, 0, 0, 0, 0, time.Local)
				return &t
			}(),
			wantItems: []ParsedItem{
				{
					Date:        time.Date(currentYear, time.January, 1, 0, 0, 0, 0, time.Local),
					Description: "TRSF E-BANKING DB 01/01",
					Debit:       500000.00,
					Credit:      0,
					Balance:     10000000.00,
				},
				{
					Date:        time.Date(currentYear, time.January, 5, 0, 0, 0, 0, time.Local),
					Description: "TRSF E-BANKING CR 05/01",
					Debit:       0,
					Credit:      1200000.00,
					Balance:     11200000.00,
				},
				{
					Date:        time.Date(currentYear, time.January, 15, 0, 0, 0, 0, time.Local),
					Description: "PAYMENT ELECTRICITY",
					Debit:       350000.00,
					Credit:      0,
					Balance:     10850000.00,
				},
				{
					Date:        time.Date(currentYear, time.January, 25, 0, 0, 0, 0, time.Local),
					Description: "SALARY JANUARY 2025",
					Debit:       0,
					Credit:      15000000.00,
					Balance:     25850000.00,
				},
			},
			wantErr: false,
		},
		{
			name: "text with installment keyword CICILAN",
			text: `01/02  CICILAN KPR BCA 12345          2,500,000.00 DB  23,350,000.00
01/02  CICILAN KENDARAAN               1,800,000.00 DB  21,550,000.00
`,
			wantItemCount: 2,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(currentYear, time.February, 1, 0, 0, 0, 0, time.Local),
					Description: "CICILAN KPR BCA 12345",
					Debit:       2500000.00,
					Credit:      0,
					Balance:     23350000.00,
				},
				{
					Date:        time.Date(currentYear, time.February, 1, 0, 0, 0, 0, time.Local),
					Description: "CICILAN KENDARAAN",
					Debit:       1800000.00,
					Credit:      0,
					Balance:     21550000.00,
				},
			},
			wantErr: false,
		},
		{
			name:            "empty text input",
			text:            "",
			wantItemCount:   0,
			wantPeriodStart: nil,
			wantPeriodEnd:   nil,
			wantErr:         false,
		},
		{
			name: "text with no matching transactions",
			text: `REKENING : 1234567890
PT COMPANY NAME
PERIODE : 01/01/2025 - 31/01/2025

TANGGAL    KETERANGAN                      MUTASI        SALDO
This is some random text that does not match
Another line without transaction data
Summary: Total Debit 1,500,000.00 Total Credit 2,000,000.00
`,
			wantItemCount:   0,
			wantPeriodStart: nil,
			wantPeriodEnd:   nil,
			wantErr:         false,
		},
		{
			name: "period and date extraction across months",
			text: `28/12  PURCHASE TOKOPEDIA              150,000.00 DB  9,850,000.00
05/01  TRSF E-BANKING CR 05/01         3,000,000.00 CR  12,850,000.00
15/01  ATM WITHDRAWAL                   500,000.00 DB  12,350,000.00
`,
			wantItemCount: 3,
			wantPeriodStart: func() *time.Time {
				t := time.Date(currentYear, time.January, 5, 0, 0, 0, 0, time.Local)
				return &t
			}(),
			wantPeriodEnd: func() *time.Time {
				t := time.Date(currentYear, time.December, 28, 0, 0, 0, 0, time.Local)
				return &t
			}(),
			wantErr: false,
		},
		{
			name: "single debit transaction",
			text: `10/03  TRANSFER TO 7890123456          1,000,000.00 DB  5,000,000.00
`,
			wantItemCount: 1,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(currentYear, time.March, 10, 0, 0, 0, 0, time.Local),
					Description: "TRANSFER TO 7890123456",
					Debit:       1000000.00,
					Credit:      0,
					Balance:     5000000.00,
				},
			},
			wantErr: false,
		},
		{
			name: "single credit transaction",
			text: `20/06  INCOMING TRANSFER FROM 111222   7,500,000.00 CR  12,500,000.00
`,
			wantItemCount: 1,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(currentYear, time.June, 20, 0, 0, 0, 0, time.Local),
					Description: "INCOMING TRANSFER FROM 111222",
					Debit:       0,
					Credit:      7500000.00,
					Balance:     12500000.00,
				},
			},
			wantErr: false,
		},
		{
			name:            "whitespace only text",
			text:            "   \n  \n\t\n  ",
			wantItemCount:   0,
			wantPeriodStart: nil,
			wantPeriodEnd:   nil,
			wantErr:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parser := &BCAParser{}
			result, err := parser.Parse(tc.text)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Len(t, result.Items, tc.wantItemCount)

			if tc.wantPeriodStart != nil {
				require.NotNil(t, result.PeriodStart, "PeriodStart should not be nil")
				assert.Equal(t, *tc.wantPeriodStart, *result.PeriodStart)
			} else if tc.wantItemCount == 0 {
				assert.Nil(t, result.PeriodStart)
			}

			if tc.wantPeriodEnd != nil {
				require.NotNil(t, result.PeriodEnd, "PeriodEnd should not be nil")
				assert.Equal(t, *tc.wantPeriodEnd, *result.PeriodEnd)
			} else if tc.wantItemCount == 0 {
				assert.Nil(t, result.PeriodEnd)
			}

			if tc.wantItems != nil {
				require.Len(t, result.Items, len(tc.wantItems))
				for i, want := range tc.wantItems {
					got := result.Items[i]
					assert.Equal(t, want.Date, got.Date, "item[%d] Date mismatch", i)
					assert.Equal(t, want.Description, got.Description, "item[%d] Description mismatch", i)
					assert.InDelta(t, want.Debit, got.Debit, 0.01, "item[%d] Debit mismatch", i)
					assert.InDelta(t, want.Credit, got.Credit, 0.01, "item[%d] Credit mismatch", i)
					assert.InDelta(t, want.Balance, got.Balance, 0.01, "item[%d] Balance mismatch", i)
				}
			}
		})
	}
}
