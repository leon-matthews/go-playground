package creditcard_test

import (
	"testing"

	"creditcard"
)

func TestNew(t *testing.T) {
	t.Parallel()
	want := "1234567890"
	cc, err := creditcard.New(want)
	if err != nil {
		t.Fatal(err)
	}
	got := cc.Number()
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestNewError(t *testing.T) {
	t.Parallel()
	_, err := creditcard.New("")
	if err == nil {
		t.Fatal("want error for invalid card number, got nil")
	}
}
