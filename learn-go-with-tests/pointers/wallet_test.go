package main

import (
	"errors"
	"testing"
)

func TestWallet(t *testing.T) {
	t.Run("deposit", func(t *testing.T) {
		w := Wallet{}
		w.Deposit(10)
		assertBalance(t, w, 10)
	})

	t.Run("withdrawal", func(t *testing.T) {
		w := Wallet{balance: Bitcoin(20)}
		err := w.Withdraw(10)
		assertNoError(t, err)
		assertBalance(t, w, 10)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		startingBalance := Bitcoin(10)
		w := Wallet{balance: startingBalance}
		err := w.Withdraw(20)
		assertErrorIs(t, err, ErrInsufficientFunds)
		assertBalance(t, w, startingBalance)
	})

	t.Run("string", func(t *testing.T) {
		btc := Bitcoin(10)
		got := btc.String()
		want := "10 BTC"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func assertBalance(t testing.TB, wallet Wallet, want Bitcoin) {
	t.Helper()
	got := wallet.Balance()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func assertErrorIs(t testing.TB, got error, want error) {
	t.Helper()
	if got == nil {
		t.Errorf("expected an error: %q", want)
	}

	if errors.Is(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertNoError(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Fatal("got an error but didn't want one")
	}
}
