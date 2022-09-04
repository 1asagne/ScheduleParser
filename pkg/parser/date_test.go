// Package parser implements structs and functions to parse events from pdf content.

package parser

import (
	"reflect"
	"testing"
	"time"

	"github.com/ledongthuc/pdf"
)

func TestEventDate_normalize(t *testing.T) {
	t.Run("FutureDate", func(t *testing.T) {
		eventDate := EventDate{time.Date(0, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(0, 5, 1, 0, 0, 0, 0, time.UTC), "once"}
		date := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		eventDate.normalize(date)
		if year := eventDate.Start.Year(); year != 2000 {
			t.Errorf("eventDate.Start.Year() = %d, want %d", year, 2000)
		}
	})

	t.Run("PastDate", func(t *testing.T) {
		eventDate := EventDate{time.Date(0, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(0, 5, 1, 0, 0, 0, 0, time.UTC), "once"}
		date := time.Date(2000, 6, 1, 0, 0, 0, 0, time.UTC)
		eventDate.normalize(date)
		if year := eventDate.Start.Year(); year != 2001 {
			t.Errorf("eventDate.Start.Year() = %d, want %d", year, 2001)
		}
	})
}

func Test_parseDates(t *testing.T) {
	type args struct {
		raw   *RawEvent
		shift int
	}

	loc := time.FixedZone("UTC+3", 3*60*60)
	initialDate := time.Date(2000, 8, 20, 0, 0, 0, 0, loc)

	tests := []struct {
		name    string
		args    args
		want    []EventDate
		want1   int
		wantErr bool
	}{
		{
			"FrequencyEvery",
			args{
				&RawEvent{"Title. Teacher. Type. Location. [05.09-05.12 к.н.]", pdf.Point{X: 46, Y: 0}, initialDate},
				0,
			},
			[]EventDate{
				{time.Date(2000, 9, 5, 8, 30, 0, 0, loc), time.Date(2000, 12, 5, 10, 10, 0, 0, loc), "every"},
			},
			32,
			false,
		},
		{
			"FrequencyOnce",
			args{
				&RawEvent{"Title. Teacher. Type. Subgroup. Location. [05.12, 19.12]", pdf.Point{X: 233, Y: 0}, initialDate},
				1,
			},
			[]EventDate{
				{time.Date(2000, 12, 5, 12, 20, 0, 0, loc), time.Date(2000, 12, 5, 15, 50, 0, 0, loc), "once"},
				{time.Date(2000, 12, 19, 12, 20, 0, 0, loc), time.Date(2000, 12, 19, 15, 50, 0, 0, loc), "once"},
			},
			42,
			false,
		},
		{
			"FrequencyThroughout",
			args{
				&RawEvent{"Title. Teacher. Type. Subgroup. Location. [26.10-21.12 ч.н.]", pdf.Point{X: 420, Y: 0}, initialDate},
				1,
			},
			[]EventDate{
				{time.Date(2000, 10, 26, 16, 0, 0, 0, loc), time.Date(2000, 12, 21, 19, 30, 0, 0, loc), "throughout"},
			},
			42,
			false,
		},
		{
			"FrequencyHybrid",
			args{
				&RawEvent{"Title. Teacher. Type. Subgroup. Location. [02.09-28.10 к.н., 11.11]", pdf.Point{X: 233, Y: 0}, initialDate},
				0,
			},
			[]EventDate{
				{time.Date(2000, 9, 2, 12, 20, 0, 0, loc), time.Date(2000, 10, 28, 14, 0, 0, 0, loc), "every"},
				{time.Date(2000, 11, 11, 12, 20, 0, 0, loc), time.Date(2000, 11, 11, 14, 0, 0, 0, loc), "once"},
			},
			42,
			false,
		},
		{
			"ParseTimeError",
			args{
				&RawEvent{"Subject. Teacher. Type. Subgroup. Location. [02.09-28.10 к.н., 11.11]", pdf.Point{X: 233, Y: 0}, initialDate},
				6,
			},
			nil,
			-1,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseDates(tt.args.raw, tt.args.shift)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDates() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseDates() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
