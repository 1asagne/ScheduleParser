package parser

import (
	"reflect"
	"testing"
	"time"

	"github.com/ledongthuc/pdf"
)

func Test_parseEvent(t *testing.T) {
	type args struct {
		raw *RawEvent
	}

	loc := time.FixedZone("UTC+3", 3*60*60)
	initialDate := time.Date(2000, 8, 20, 0, 0, 0, 0, loc)

	tests := []struct {
		name    string
		args    args
		want    *Event
		wantErr bool
	}{
		{
			"WithoutSubgroup",
			args{&RawEvent{"Title. Teacher T.T. лекции. Location. [05.09-05.12 к.н.]", pdf.Point{X: 46, Y: 0}, initialDate}},
			&Event{"Title", "Teacher T.T.", "lecture", "", "Location", []EventDate{{time.Date(2000, 9, 5, 8, 30, 0, 0, loc), time.Date(2000, 12, 5, 10, 10, 0, 0, loc), "every"}}},
			false,
		},
		{
			"WithSubgroup",
			args{&RawEvent{"Title. Teacher T.T. лабораторные занятия. (Subgroup). Location. [19.09-17.10 ч.н.]", pdf.Point{X: 233, Y: 513}, initialDate}},
			&Event{"Title", "Teacher T.T.", "lab", "Subgroup", "Location", []EventDate{{time.Date(2000, 9, 19, 12, 20, 0, 0, loc), time.Date(2000, 10, 17, 15, 50, 0, 0, loc), "throughout"}}},
			false,
		},
		{
			"TypeNotFoundError",
			args{&RawEvent{"Title. Teacher T.T. Unknown. Location. [05.09-05.12 к.н.]", pdf.Point{X: 0, Y: 0}, initialDate}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEvent(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
