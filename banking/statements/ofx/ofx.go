// Package ofx reads OFX (Open Financial Exchange) bank and credit card statements.
package ofx

import (
	"bytes"
	"fmt"

	"github.com/aclindsa/ofxgo"

	"banking/common"
	"banking/statements"
)

// Format is the OFX statement format.
var Format ofxFormat

type ofxFormat struct{}

func init() {
	statements.Register(&Format)
}

func (ofxFormat) Name() string { return "ofx" }

// Detect checks whether the data looks like an OFX file.
func (ofxFormat) Detect(data []byte) error {
	if !bytes.Contains(data, []byte("OFXHEADER:")) && !bytes.Contains(data, []byte("<OFX>")) {
		return fmt.Errorf("no OFX header or <OFX> tag found")
	}
	return nil
}

// Read parses OFX data and returns transactions.
func (ofxFormat) Read(data []byte) ([]*common.Transaction, error) {
	resp, err := ofxgo.ParseResponse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parsing OFX: %w", err)
	}

	var transactions []*common.Transaction

	// Credit card statements
	for _, msg := range resp.CreditCard {
		stmt, ok := msg.(*ofxgo.CCStatementResponse)
		if !ok || stmt.BankTranList == nil {
			continue
		}
		account := stmt.CCAcctFrom.AcctID.String()
		for _, txn := range stmt.BankTranList.Transactions {
			t, err := convertTransaction(txn, account)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, t)
		}
	}

	// Bank statements
	for _, msg := range resp.Bank {
		stmt, ok := msg.(*ofxgo.StatementResponse)
		if !ok || stmt.BankTranList == nil {
			continue
		}
		account := stmt.BankAcctFrom.AcctID.String()
		for _, txn := range stmt.BankTranList.Transactions {
			t, err := convertTransaction(txn, account)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, t)
		}
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found in OFX data")
	}

	return transactions, nil
}

func convertTransaction(txn ofxgo.Transaction, account string) (*common.Transaction, error) {
	date := txn.DtPosted.Time
	amount, _ := txn.TrnAmt.Float64()
	name := txn.Name.String()
	memo := txn.Memo.String()
	combined := name
	if memo != "" {
		combined = memo + " " + name
	}
	details := common.CleanString(combined)

	return &common.Transaction{
		Date:      date,
		Processed: date,
		Account:   account,
		Details:   details,
		Amount:    amount,
	}, nil
}
