package web

import (
	"testing"

	regexp "github.com/dlclark/regexp2"
)

func TestEmailRegex(t *testing.T) {
	emailExp := regexp.MustCompile(emailRegex, regexp.None)
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "right email",
			email: "12223456432@qq.com",
			want:  true,
		},
		{
			name:  "wrong email",
			email: "1234567@qq",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := emailExp.MatchString(tt.email)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if ok != tt.want {
				t.Errorf("ok != tt.want. ok : %v, tt.want : %v", ok, tt.want)
			}
		})
	}
}

func TestPassRegex(t *testing.T) {
	passExp := regexp.MustCompile(passRegex, regexp.None)
	tests := []struct {
		name string
		pass string
		want bool
	}{
		{
			name: "right password",
			//pass: "1E#3fg4et",
			pass: "Ew333W#23fget",
			want: true,
		},
		{
			name: "wrong password",
			pass: "1234",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := passExp.MatchString(tt.pass)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if ok != tt.want {
				t.Errorf("ok != tt.want. ok : %v, tt.want : %v", ok, tt.want)
			}
		})
	}
}
