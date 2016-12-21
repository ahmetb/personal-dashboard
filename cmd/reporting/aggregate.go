package main

import "time"

type record struct {
	Date         time.Time `json:"key"`
	Value        float64   `json:"data"`
	Interpolated bool      `json:"fake"`
}

type aggregateFunc func([]record) []record

var (
	aggregators = map[string]aggregateFunc{
		"linear_interpolation": linearInterpolation,
		"zero_fill_days":       zeroFillMissingDays,
	}
)

// linearInterpolation completes records with missing (0) data field using
// linear polation from neighboring fields. Works only if first and last data
// points have valid data.
func linearInterpolation(in []record) []record {
	lastData := -1
	interpolate := false
	for i := 0; i < len(in); i++ {
		if in[i].Value == 0 {
			interpolate = true
		} else {
			if interpolate && lastData != -1 {
				loVal := in[lastData].Value
				hiVal := in[i].Value
				missing := i - lastData - 1
				incr := (hiVal - loVal) / float64(missing+1)

				for j := 1; j <= missing; j++ {
					in[lastData+j].Value = loVal + (incr * float64(j))
					in[lastData+j].Interpolated = true
				}
			}
			interpolate = false
			lastData = i
		}
	}
	return in
}

// zeroFillMissingDays takes a record list chronically ordered by date and adds
// missing days in between with zero value.
func zeroFillMissingDays(in []record) (out []record) {
	var prev time.Time
	for _, v := range in {
		if !prev.IsZero() {
			diffDays := int(v.Date.Sub(prev) / time.Hour / 24)
			if diffDays > 1 {
				for i := 1; i < diffDays; i++ {
					out = append(out, record{
						Date:  prev.Add(time.Hour * 24 * time.Duration(i)),
						Value: 0})
				}
			}
		}
		out = append(out, v)
		prev = v.Date
	}
	return out
}
