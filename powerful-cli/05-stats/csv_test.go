package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSV2Float(t *testing.T) {
	t.Parallel()
	csvData := `IP Address,Requests,Response Time
192.168.0.199,2056,236
192.168.0.88,899,220
192.168.0.199,3054,226
192.168.0.100,4133,218
192.168.0.199,950,238
`
	testCases := map[string]struct {
		column  int
		want    []float64
		wantErr error
		reader  io.Reader
	}{
		"Column 2": {
			column:  2,
			want:    []float64{2056, 899, 3054, 4133, 950},
			wantErr: nil,
			reader:  bytes.NewBufferString(csvData),
		},
		"Column 3": {
			column:  3,
			want:    []float64{236, 220, 226, 218, 238},
			wantErr: nil,
			reader:  bytes.NewBufferString(csvData),
		},
		"Error Read": {
			column:  1,
			want:    nil,
			wantErr: iotest.ErrTimeout,
			reader:  iotest.ErrReader(iotest.ErrTimeout),
		},
		"Error Not Number": {
			column:  1,
			want:    nil,
			wantErr: ErrNotNumber,
			reader:  bytes.NewBufferString(csvData),
		},
		"Error Invalid Column": {
			column:  4,
			want:    nil,
			wantErr: ErrInvalidColumn,
			reader:  bytes.NewBufferString(csvData),
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := csv2Float(tt.reader, tt.column)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestOperations(t *testing.T) {
	t.Parallel()
	data := [][]float64{
		{10, 20, 15, 30, 45, 50, 100, 30},
		{5.5, 8, 2.2, 9.75, 8.45, 3, 2.5, 10.25, 4.75, 6.1, 7.67, 12.287, 5.47},
		{-10, -20},
		{102, 37, 44, 57, 67, 129},
	}

	testCases := []struct {
		name      string
		operation statsFunc
		want      []float64
	}{
		{"Sum", sum, []float64{300, 85.927, -30, 436}},
		{"Mean", mean, []float64{37.5, 6.609769230769231, -15, 72.666666666666666}},
	}

	for _, tt := range testCases {
		for i, want := range tt.want {
			name := fmt.Sprintf("%sData%d", tt.name, i)
			t.Run(name, func(t *testing.T) {
				got := tt.operation(data[i])
				assert.Equal(t, want, got)
			})
		}
	}
}
