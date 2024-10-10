package pointers_errors

import (
	"errors"
	"fmt"
)

type Bitcoin int

var ErrorInsufficientFunds = errors.New("insufficient funds")

func (b Bitcoin) String() string {
	return fmt.Sprintf("%d BTC", b)
}

type Wallet struct {
	balance Bitcoin
}

// Fetch the wallet's balance
func (w *Wallet) Balance() Bitcoin {
	return w.balance
}

// Add the given amount to the wallet's balance
func (w *Wallet) Deposit(amount Bitcoin) {
	w.balance += amount
}

func (w *Wallet) String() string {
	return w.balance.String()
}

// Remove and return the given amount from the wallet's balance
func (w *Wallet) Withdraw(amount Bitcoin) error {
	if amount > w.balance {
		return ErrorInsufficientFunds
	}

	w.balance -= amount
	return nil
}
