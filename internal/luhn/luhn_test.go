package luhn

import "testing"

func TestCalculateLuhn(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "success test #1",
			args: args{
				number: 1234567890,
			},
			want: 3,
		},
		{
			name: "success test #2",
			args: args{
				number: 12345678903,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateLuhn(tt.args.number); got != tt.want {
				t.Errorf("CalculateLuhn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValid(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success test #1",
			args: args{
				number: 12345678903,
			},
			want: true,
		},
		{
			name: "success test #2",
			args: args{
				number: 123456789031,
			},
			want: true,
		},
		{
			name: "fail test #1",
			args: args{
				number: 12345678902,
			},
			want: false,
		},
		{
			name: "fail test #2",
			args: args{
				number: 123456789033,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Valid(tt.args.number); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checksum(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "success test #1",
			args: args{
				number: 12345678903,
			},
			want: 9,
		},
		{
			name: "success test #2",
			args: args{
				number: 123456789031,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checksum(tt.args.number); got != tt.want {
				t.Errorf("checksum() = %v, want %v", got, tt.want)
			}
		})
	}
}
