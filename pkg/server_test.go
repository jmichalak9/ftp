package ftp

import "testing"

type handlerTest struct {
	title    string
	c        client
	argv     string
	expected string
}

// Handler tests only check for returned reply code.
func TestHandleSYST(t *testing.T) {
	tests := []handlerTest{
		{
			title:    "SYST without arguments",
			argv:     "",
			expected: "215",
		},
	}
	for _, test := range tests {
		result := handleSYST(&test.c, test.argv)
		replyCode := replyCodeFromResult(result)
		if replyCode != test.expected {
			t.Errorf("failed, got %v, expected %v", replyCode, test.expected)
		}
	}
}

func replyCodeFromResult(result string) string {
	if len(result) < 3 {
		return result
	}
	return result[0:3]
}
