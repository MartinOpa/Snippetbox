package main

import (
	"testing"
	"time"
)

// func Test<FunctionName>(t *testing.T)
func TestHumanDate(t *testing.T) {
	// entry time value
	// tm := time.Date(2022, 12, 28, 10, 0, 0, 0, time.UTC)
	// exp := "28 Dec 2022 at 10:00"
	// converted time value
	// hd := humanDate(tm)

	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2022, 12, 28, 10, 0, 0, 0, time.UTC),
			want: "28 Dec 2022 at 10:00",
		},
		{
			name: "CET",
			tm:   time.Date(2022, 12, 28, 10, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "28 Dec 2022 at 10:00",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)

			if hd != tt.want {
				t.Errorf("want %q; got %q", tt.want, hd)
			}
		})
	}
}
