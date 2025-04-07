package template

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDefaultTemplate_Execute(t1 *testing.T) {
	type fields struct {
		name  string
		text  string
		goFmt bool
	}
	type args struct {
		data any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *bytes.Buffer
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				name:  "test",
				text:  "text",
				goFmt: true,
			},
			args: args{
				data: nil,
			},
			want:    bytes.NewBufferString("text"),
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				name:  "test",
				text:  "text",
				goFmt: false,
			},
			args: args{
				data: nil,
			},
			want:    bytes.NewBufferString("text"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := NewTemplate().Parse(tt.fields.text).GoFmt(tt.fields.goFmt)
			got, err := t.Execute(tt.args.data)
			if (err != nil) != tt.wantErr {
				t1.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Execute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
