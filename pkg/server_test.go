package ftp

import "testing"

type handlerTest struct {
	title    string
	argv     string
	expected string
}

func TestHandleSYST(t *testing.T) {
	tests := []handlerTest{
		{
			title:    "SYST without arguments",
			argv:     "",
			expected: "215",
		},
	}
	for _, test := range tests {
		result := handleSYST(test.client, test.argv)
		if result != test.expected {
			t.Errorf("failed, got %v, expected %v", result, test.expected)
		}
	}
}
