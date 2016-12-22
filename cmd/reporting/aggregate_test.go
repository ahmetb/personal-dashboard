package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_linearInterpolation(t *testing.T) {
	tests := []struct {
		in  []record
		out []record
	}{
		// nil arr
		{in: nil,
			out: nil},

		// all zeros
		{in: []record{r(0)},
			out: []record{r(0)}},

		// no zeros
		{in: []record{r(3), r(4), r(5)},
			out: []record{r(3), r(4), r(5)}},

		// zeroes at the end, not interpolated
		{in: []record{r(3), r(4), r(0), r(0)},
			out: []record{r(3), r(4), r(0), r(0)}},

		// zeros at the beginning, not interpolated
		{in: []record{r(0), r(0), r(4), r(5)},
			out: []record{r(0), r(0), r(4), r(5)}},

		// interpolate towards up
		{in: []record{r(10), r(0), r(0), r(40)},
			out: []record{r(10), ri(20), ri(30), r(40)}},

		// interpolate towards down
		{in: []record{r(6), r(0), r(0), r(0), r(5)},
			out: []record{r(6), ri(5.75), ri(5.5), ri(5.25), r(5)}},
	}
	for i, tt := range tests {
		got := linearInterpolation(tt.in)
		require.Equal(t, tt.out, got, "case:#%d", i)
	}
}

func Test_zeroFillMissingDays(t *testing.T) {
	tests := []struct {
		in  []record
		out []record
	}{
		// nil arr
		{in: nil,
			out: nil},

		// single record
		{in: []record{d(2016, 2, 27)},
			out: []record{d(2016, 2, 27)}},

		// no missing records
		{in: []record{d(2016, 2, 28), d(2016, 2, 29), d(2016, 2, 30)},
			out: []record{d(2016, 2, 28), d(2016, 2, 29), d(2016, 2, 30)}},

		// fills missing records
		{in: []record{d(2016, 2, 27), d(2016, 3, 2)},
			out: []record{d(2016, 2, 27), d(2016, 2, 28), d(2016, 2, 29), d(2016, 3, 1), d(2016, 3, 2)}},
	}
	for i, tt := range tests {
		got := zeroFillMissingDays(tt.in)
		require.Equal(t, tt.out, got, "case:#%d", i)
	}
}

func r(v float64) record  { return record{Value: v} }
func ri(v float64) record { return record{Value: v, Interpolated: true} }

func d(y int, m time.Month, d int) record {
	return record{Date: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}
