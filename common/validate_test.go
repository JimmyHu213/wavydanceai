package common

import "testing"

func TestIsPasswordComplexEnough(t *testing.T) {
	cases := []struct {
		name string
		pw   string
		want bool
	}{
		{"empty", "", false},
		{"too short", "abc12", false},
		{"min ok letter+digit", "abcdefg1", true},
		{"max ok 24 letter+digit", "abcdefghijklmnopqrstuvw1", true},
		{"too long 25", "abcdefghijklmnopqrstuvwx1", false},
		{"letters only at min", "abcdefgh", false},
		{"digits only at min", "12345678", false},
		{"non-ascii letter + digit", "пароль123", true},
		{"symbols + letter + digit", "Aa1!_-_-", true},
		{"symbols only", "!@#$%^&*", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsPasswordComplexEnough(tc.pw)
			if got != tc.want {
				t.Fatalf("IsPasswordComplexEnough(%q) = %v, want %v", tc.pw, got, tc.want)
			}
		})
	}
}
