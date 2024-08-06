package main

import (
	"snippetbox.i4o.dev/internal/assert"
	"testing"
	"time"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2024, 8, 6, 7, 30, 0, 0, time.UTC),
			want: "06 Aug 2024 at 07:30",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2024, 8, 6, 7, 30, 0, 0, time.FixedZone("IST", 5*60*60+30*60)),
			want: "06 Aug 2024 at 02:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)

			assert.Equal(t, hd, tt.want)
		})
	}
}
