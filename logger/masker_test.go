package logger

import "testing"

func Test_maskImpl_Mask(t *testing.T) {
	type fields struct {
		keywords []string
	}
	type args struct {
		text string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "testcase #1",
			fields: fields{keywords: []string{"password", "email", "ccNumber"}},
			args:   args{text: `{"password"             : "usss" ,   "numb":1   , "ccNumber": 123123982347827  }`},
			want:   `{"password"             : "xxxxx",   "numb":1   , "ccNumber": "xxxxx"}`,
		},
		{
			name:   "testcase #2",
			fields: fields{keywords: []string{"password", "email", "ccNumber"}},
			args:   args{text: `{"password":"cok","ccNumber":123123982347827}`},
			want:   `{"password":"xxxxx","ccNumber":"xxxxx"}`,
		},
		{
			name:   "testcase #3",
			fields: fields{keywords: []string{"password", "email", "ccNumber"}},
			args:   args{text: `user=test&password=king`},
			want:   `user=test&password="xxxxx"`,
		},
		{
			name:   "testcase #4",
			fields: fields{keywords: []string{"password", "email", "ccNumber"}},
			args:   args{text: `<password>password</password>`},
			want:   `<password>"xxxxx"</password>`,
		},
		{
			name:   "testcase #5",
			fields: fields{keywords: []string{"password", "email", "ccNumber"}},
			args:   args{text: `password: testtsts, u   : user`},
			want:   `password: "xxxxx", u   : user`,
		},
		{
			name:   "testcase #6",
			fields: fields{keywords: []string{"password", "email", "ccNumber"}},
			args: args{text: `data: {
		email: "candi@golang.id",
		password: 2384823748273847287981739817389127389712983719873981273127317
}`},
			want: `data: {
		email: "xxxxx",
		password: "xxxxx"
}`,
		},
		{
			name:   "testcase #7",
			fields: fields{keywords: []string{"password", "email", "ccNumber"}},
			args:   args{text: `{"password"		:"usss asddasd"}`},
			want:   `{"password"		:"xxxxx"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMasker(tt.fields.keywords...)
			if got := r.Mask(tt.args.text); got != tt.want {
				t.Errorf("maskImpl.Mask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkMasker(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewMasker("password", "email", "ccNumber", "numb")
		r.Mask(`{"password"             : "usss" ,   "numb":1   , "ccNumber": 123123982347827  }`)
	}
}
