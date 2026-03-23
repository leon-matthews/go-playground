package statements

import (
	"fmt"
	"testing"

	"banking/common"
)

type mockFormat struct {
	name     string
	detectFn func(data []byte) error
}

func (m mockFormat) Name() string                               { return m.name }
func (m mockFormat) Detect(data []byte) error                   { return m.detectFn(data) }
func (m mockFormat) Read([]byte) ([]*common.Transaction, error) { return nil, nil }

func TestDetect(t *testing.T) {
	saved := formats
	t.Cleanup(func() { formats = saved })

	formats = nil
	Register(mockFormat{
		name: "testbank",
		detectFn: func(data []byte) error {
			if string(data) == "testbank-data" {
				return nil
			}
			return fmt.Errorf("not testbank format")
		},
	})

	t.Run("detects matching format", func(t *testing.T) {
		got, err := Detect([]byte("testbank-data"))
		if err != nil {
			t.Fatal(err)
		}
		if got.Name() != "testbank" {
			t.Errorf("detected format = %q, want testbank", got.Name())
		}
	})

	t.Run("unrecognised format", func(t *testing.T) {
		_, err := Detect([]byte("unknown-data"))
		if err == nil {
			t.Fatal("expected error for unrecognised format")
		}
	})
}

func TestGet(t *testing.T) {
	saved := formats
	t.Cleanup(func() { formats = saved })

	formats = nil
	Register(mockFormat{name: "mybank", detectFn: func([]byte) error { return fmt.Errorf("no match") }})

	t.Run("known format", func(t *testing.T) {
		f, ok := Get("mybank")
		if !ok {
			t.Fatal("expected to find mybank format")
		}
		if f.Name() != "mybank" {
			t.Errorf("name = %q, want mybank", f.Name())
		}
	})

	t.Run("unknown format", func(t *testing.T) {
		_, ok := Get("nonexistent")
		if ok {
			t.Fatal("expected not to find nonexistent format")
		}
	})
}

func TestNames(t *testing.T) {
	saved := formats
	t.Cleanup(func() { formats = saved })

	formats = nil
	Register(mockFormat{name: "bank_a", detectFn: func([]byte) error { return fmt.Errorf("no match") }})
	Register(mockFormat{name: "bank_b", detectFn: func([]byte) error { return fmt.Errorf("no match") }})

	names := Names()
	if len(names) != 2 {
		t.Fatalf("got %d names, want 2", len(names))
	}
	if names[0] != "bank_a" || names[1] != "bank_b" {
		t.Errorf("names = %v, want [bank_a bank_b]", names)
	}
}
