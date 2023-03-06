package filter

import (
	"golang.org/x/exp/slices"
)

// describes any function which can receive a log label
// and the actual log line and based on this information
// decides to include the item or not
type Func func(string, []byte) bool

func Default(label string, data []byte) bool {
	return true
}

// WithInclude returns true if at least one field passed
// the filter else false
// func WithInclude(streams ...string) Func {
// 	return func(label string, data []byte) bool {

// 		var include bool

// 		for _, stream := range streams {
// 			include = stream == label
// 		}

// 		return include
// 	}
// }

type FilterFunc func(item string, label string, data []byte) bool

type Filter struct {
	fields []string
	fn     FilterFunc
}

func New(filter FilterFunc, fields ...string) *Filter {
	return &Filter{
		fields: fields,
		fn:     filter,
	}
}

func (filter Filter) Test(label string, data []byte) bool {

	for _, field := range filter.fields {
		if filter.fn(field, label, data) {
			return true
		}
	}

	return false
}

func (filter *Filter) Append(fields ...string) {

	for _, field := range fields {
		if slices.Contains(filter.fields, field) {
			continue
		}
		filter.fields = append(filter.fields, field)
	}

}

func (filter *Filter) Remove(field string) {

	var offset int = -1 // invalid index

	for i, f := range filter.fields {
		if field == f {
			offset = i
			break
		}
	}

	// item to remove was not found; Nop
	if offset < 0 {
		return
	}

	filter.fields = append(filter.fields[:offset], filter.fields[offset+1:]...)
}
