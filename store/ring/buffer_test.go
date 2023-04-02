package ring

import (
	"fmt"
	"testing"
)

func TestRangeNoOverflow(t *testing.T) {

	buffer := New(12)

	for i := 0; i < 12; i++ { // overflow by 18-12 => 6 items
		buffer.Insert(Item{
			Raw: fmt.Sprintf("Line-%d", i+1),
		})
	}

	tt := []struct {
		name  string
		start int
		size  int
		want  []Item
	}{
		{
			name:  "range all",
			start: 0,
			size:  12,
			want: []Item{
				{Raw: "Line-1"},
				{Raw: "Line-2"},
				{Raw: "Line-3"},
				{Raw: "Line-4"},
				{Raw: "Line-5"},
				{Raw: "Line-6"},
				{Raw: "Line-7"},
				{Raw: "Line-8"},
				{Raw: "Line-9"},
				{Raw: "Line-10"},
				{Raw: "Line-11"},
				{Raw: "Line-12"},
			},
		},
		{
			name:  "range middel part",
			start: 4,
			size:  8,
			want: []Item{
				{Raw: "Line-5"},
				{Raw: "Line-6"},
				{Raw: "Line-7"},
				{Raw: "Line-8"},
				{Raw: "Line-9"},
			},
		},
	}

	for _, tc := range tt {
		res := buffer.Range(tc.start, tc.size)

		for i, item := range tc.want {
			if item.Raw != res[i].Raw {
				t.Fatalf("[%s] wanted item: %s; got item: %s", tc.name, item.Raw, res[i].Raw)
			}
		}
	}
}

func TestRangeOverflow(t *testing.T) {

	buffer := New(12)

	for i := 0; i < 18; i++ { // overflow by 18-12 => 6 items
		buffer.Insert(Item{
			Raw: fmt.Sprintf("Line-%d", i+1),
		})
	}

	tt := []struct {
		name  string
		start int
		size  int
		want  []Item
	}{
		{
			name:  "range with no overflow part",
			start: 6,
			size:  10,
			want: []Item{
				{Raw: "Line-7"},
				{Raw: "Line-8"},
				{Raw: "Line-9"},
				{Raw: "Line-10"},
			},
		},
		{
			name:  "range with overflow part",
			start: 10,
			size:  4,
			want: []Item{
				{Raw: "Line-11"},
				{Raw: "Line-12"},
				{Raw: "Line-13"},
				{Raw: "Line-14"},
			},
		},
	}

	for _, tc := range tt {
		res := buffer.Range(tc.start, tc.size)

		for i, item := range tc.want {
			if item.Raw != res[i].Raw {
				t.Fatalf("[%s] wanted item: %s; got item: %s", tc.name, item.Raw, res[i].Raw)
			}
		}
	}
}
