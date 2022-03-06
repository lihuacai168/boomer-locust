package main

import (
	"reflect"
	"testing"
)

func TestHandleReplaceBody(t *testing.T) {
	type args struct {
		dataType  string
		rawData   []byte
		row       []string
		replaceKV map[string]int
	}
	replaceKV := map[string]int{
		"$a": 0,
		"$b": 1,
		"$c": 2,
	}
	rawData := []byte(`{"a": "$a", "b": "$b"}`)
	row := []string{"100", "101", "102"}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test1-int",
			args: args{
				dataType:  "int",
				rawData:   rawData,
				row:       row,
				replaceKV: replaceKV,
			},
			want: []byte(`{"a":100,"b":101}`),
		},
		{
			name: "test2-string",
			args: args{
				dataType:  "string",
				rawData:   rawData,
				row:       row,
				replaceKV: replaceKV,
			},
			want: []byte(`{"a":"100","b":"101"}`),
		},
		{
			name: "test3-intArray",
			args: args{
				dataType:  "intArray",
				rawData:   rawData,
				row:       row,
				replaceKV: replaceKV,
			},
			want: []byte(`{"a":[100],"b":[101]}`),
		},
		{
			name: "test4-interface",
			args: args{
				dataType:  "interface",
				rawData:   []byte(`{"a": "$a", "b": "$b", "d": "$d", "e": "e"}`),
				row:       row,
				replaceKV: replaceKV,
			},
			want: []byte(`{"a":100,"b":101,"d":"$d","e":"e"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HandleReplaceBody(tt.args.rawData, tt.args.dataType, tt.args.row, tt.args.replaceKV); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleReplaceBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
