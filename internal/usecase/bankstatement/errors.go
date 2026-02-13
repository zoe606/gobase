package bankstatement

import "errors"

var (
	ErrUnsupportedBank  = errors.New("unsupported bank")
	ErrInvalidPDF       = errors.New("invalid or corrupted PDF file")
	ErrParsingFailed    = errors.New("failed to parse bank statement")
	ErrDecryptionFailed = errors.New("failed to decrypt PDF - wrong password")
)
