// Package scheduleparser implements structs and functions to parse events from pdf content.

package scheduleparser

import (
	"reflect"
	"testing"
	"time"

	"github.com/ledongthuc/pdf"
)

func Test_parseTime(t *testing.T) {
	type args struct {
		raw   *RawEvent
		shift int
	}
	tests := []struct {
		name    string
		args    args
		want    *EventTime
		wantErr bool
	}{
		{
			"ZeroWithoutShift",
			args{
				&RawEvent{"", pdf.Point{X: 46, Y: 0}, time.Time{}},
				0,
			},
			&eventTimes[0],
			false,
		},
		{
			"ZeroWithShift",
			args{
				&RawEvent{"", pdf.Point{X: 46, Y: 0}, time.Time{}},
				1,
			},
			&EventTime{Clock{8, 30}, Clock{12, 0}},
			false,
		},
		{
			"SeventhWithoutShift",
			args{
				&RawEvent{"", pdf.Point{X: 700, Y: 0}, time.Time{}},
				0,
			},
			&eventTimes[7],
			false,
		},
		{
			"ShiftError",
			args{
				&RawEvent{"", pdf.Point{X: 46, Y: 0}, time.Time{}},
				8,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTime(tt.args.raw, tt.args.shift)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
