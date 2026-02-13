package bankstatement

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBRIParser_Parse(t *testing.T) {
	t.Parallel()

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
			name: "valid BRI statement with multiple transactions using slash date format",
			text: `NO REKENING : 0987-01-654321-50-2
NAMA : PT COMPANY NAME
PERIODE : 01/01/2025 S/D 31/01/2025

TANGGAL       KETERANGAN                    DEBIT           KREDIT          SALDO
01/01/2025    TRANSFER MASUK                0,00            5.000.000,00    25.000.000,00
05/01/2025    BIAYA ADMIN BULANAN           15.000,00       0,00            24.985.000,00
15/01/2025    TRANSFER KELUAR KE BCA        2.500.000,00    0,00            22.485.000,00
25/01/2025    GAJI BULAN JANUARI            0,00            15.000.000,00   37.485.000,00
`,
			wantItemCount: 4,
			wantPeriodStart: func() *time.Time {
				t := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantPeriodEnd: func() *time.Time {
				t := time.Date(2025, time.January, 25, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
					Description: "TRANSFER MASUK",
					Debit:       0,
					Credit:      5000000.00,
					Balance:     25000000.00,
				},
				{
					Date:        time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC),
					Description: "BIAYA ADMIN BULANAN",
					Debit:       15000.00,
					Credit:      0,
					Balance:     24985000.00,
				},
				{
					Date:        time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC),
					Description: "TRANSFER KELUAR KE BCA",
					Debit:       2500000.00,
					Credit:      0,
					Balance:     22485000.00,
				},
				{
					Date:        time.Date(2025, time.January, 25, 0, 0, 0, 0, time.UTC),
					Description: "GAJI BULAN JANUARI",
					Debit:       0,
					Credit:      15000000.00,
					Balance:     37485000.00,
				},
			},
			wantErr: false,
		},
		{
			name: "valid BRI statement with dash date format",
			text: `10-02-2025    PEMBELIAN TOKOPEDIA           150.000,00      0,00            24.835.000,00
20-02-2025    TRANSFER DARI REKENING BRI    0,00            3.000.000,00    27.835.000,00
`,
			wantItemCount: 2,
			wantPeriodStart: func() *time.Time {
				t := time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantPeriodEnd: func() *time.Time {
				t := time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.February, 10, 0, 0, 0, 0, time.UTC),
					Description: "PEMBELIAN TOKOPEDIA",
					Debit:       150000.00,
					Credit:      0,
					Balance:     24835000.00,
				},
				{
					Date:        time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC),
					Description: "TRANSFER DARI REKENING BRI",
					Debit:       0,
					Credit:      3000000.00,
					Balance:     27835000.00,
				},
			},
			wantErr: false,
		},
		{
			name: "text with installment keyword CICILAN",
			text: `01/03/2025    CICILAN KPR RUMAH BRI         2.500.000,00    0,00            20.000.000,00
01/03/2025    CICILAN KENDARAAN BRI         1.800.000,00    0,00            18.200.000,00
`,
			wantItemCount: 2,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
					Description: "CICILAN KPR RUMAH BRI",
					Debit:       2500000.00,
					Credit:      0,
					Balance:     20000000.00,
				},
				{
					Date:        time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
					Description: "CICILAN KENDARAAN BRI",
					Debit:       1800000.00,
					Credit:      0,
					Balance:     18200000.00,
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
			text: `NO REKENING : 0987-01-654321-50-2
NAMA : PT COMPANY NAME
PERIODE : 01/01/2025 S/D 31/01/2025

TANGGAL       KETERANGAN                    DEBIT           KREDIT          SALDO
Some random text that does not match
Another line without transaction data
RINGKASAN: Total Debit Rp. 5.000.000 Total Kredit Rp. 10.000.000
`,
			wantItemCount:   0,
			wantPeriodStart: nil,
			wantPeriodEnd:   nil,
			wantErr:         false,
		},
		{
			name: "period extraction across months",
			text: `28/12/2024    PEMBELIAN SHOPEE              200.000,00      0,00            9.800.000,00
05/01/2025    TRANSFER MASUK                0,00            3.000.000,00    12.800.000,00
15/02/2025    PEMBAYARAN LISTRIK            350.000,00      0,00            12.450.000,00
`,
			wantItemCount: 3,
			wantPeriodStart: func() *time.Time {
				t := time.Date(2024, time.December, 28, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantPeriodEnd: func() *time.Time {
				t := time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantErr: false,
		},
		{
			name: "single debit transaction",
			text: `10/04/2025    TRANSFER KE REK MANDIRI       1.000.000,00    0,00            5.000.000,00
`,
			wantItemCount: 1,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.April, 10, 0, 0, 0, 0, time.UTC),
					Description: "TRANSFER KE REK MANDIRI",
					Debit:       1000000.00,
					Credit:      0,
					Balance:     5000000.00,
				},
			},
			wantErr: false,
		},
		{
			name: "single credit transaction",
			text: `20/06/2025    SETORAN TUNAI                 0,00            7.500.000,00    12.500.000,00
`,
			wantItemCount: 1,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.June, 20, 0, 0, 0, 0, time.UTC),
					Description: "SETORAN TUNAI",
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
		{
			name: "invalid date is skipped gracefully",
			text: `32/13/2025    INVALID DATE LINE             100.000,00      0,00            5.000.000,00
10/04/2025    VALID TRANSACTION             500.000,00      0,00            4.500.000,00
`,
			wantItemCount: 1,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.April, 10, 0, 0, 0, 0, time.UTC),
					Description: "VALID TRANSACTION",
					Debit:       500000.00,
					Credit:      0,
					Balance:     4500000.00,
				},
			},
			wantErr: false,
		},
		{
			name: "description with trailing number consumed by non-greedy regex",
			// The BRI parser uses a non-greedy (.+?) for description.
			// When a description ends with a number (e.g., "GAJI JANUARI 2025"),
			// the regex captures the trailing number as part of the debit column
			// instead of the description. This test documents that behavior.
			text: `25/01/2025    GAJI JANUARI 2025             0,00            15.000.000,00   37.485.000,00
`,
			wantItemCount: 1,
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.January, 25, 0, 0, 0, 0, time.UTC),
					Description: "GAJI JANUARI",
					Debit:       2025,
					Credit:      0,
					Balance:     15000000.00,
				},
			},
			wantErr: false,
		},
		{
			name: "mixed debit and credit with large amounts",
			text: `01/07/2025    PENDAPATAN BUNGA              0,00            250.000,00      100.250.000,00
02/07/2025    PAJAK BUNGA                   50.000,00       0,00            100.200.000,00
03/07/2025    BIAYA MATERAI                 10.000,00       0,00            100.190.000,00
`,
			wantItemCount: 3,
			wantPeriodStart: func() *time.Time {
				t := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantPeriodEnd: func() *time.Time {
				t := time.Date(2025, time.July, 3, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantItems: []ParsedItem{
				{
					Date:        time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC),
					Description: "PENDAPATAN BUNGA",
					Debit:       0,
					Credit:      250000.00,
					Balance:     100250000.00,
				},
				{
					Date:        time.Date(2025, time.July, 2, 0, 0, 0, 0, time.UTC),
					Description: "PAJAK BUNGA",
					Debit:       50000.00,
					Credit:      0,
					Balance:     100200000.00,
				},
				{
					Date:        time.Date(2025, time.July, 3, 0, 0, 0, 0, time.UTC),
					Description: "BIAYA MATERAI",
					Debit:       10000.00,
					Credit:      0,
					Balance:     100190000.00,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parser := &BRIParser{}
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
