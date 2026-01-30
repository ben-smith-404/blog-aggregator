package config

import (
	"testing"
)

type testConfigStruct struct {
	testName string
	userName string
	expect   Config
}

var tests []testConfigStruct

// TestReadConfig is used to test that the file can be read without anything being written
func TestReadConfig(t *testing.T) {
	test := testConfigStruct{
		testName: "basic read",
		expect: Config{
			DbURL: "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
		},
	}

	got, err := Read()
	if err != nil {
		t.Error(err)
	}
	if got.DbURL != test.expect.DbURL {
		t.Errorf("Error testing %v: expected %v but got %v", test.testName, test.expect, got)
	}

}

// TestSetUserConfig tests that the user name can be updated in the file
func TestSetUserConfig(t *testing.T) {
	tests = append(tests, testConfigStruct{
		testName: "basic write",
		userName: "Robin Hood",
		expect: Config{
			DbURL:           "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
			CurrentUserName: "Robin Hood",
		},
	})
	tests = append(tests, testConfigStruct{
		testName: "follow-up write",
		userName: "Friar Tuck",
		expect: Config{
			DbURL:           "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
			CurrentUserName: "Friar Tuck",
		},
	})

	config, err := Read()
	if err != nil {
		t.Error(err)
	}

	for _, test := range tests {
		err := config.SetUser(test.userName)
		if err != nil {
			t.Error(err)
		}
		got, err := Read()
		if err != nil {
			t.Error(err)
		}
		if got != test.expect {
			t.Errorf("Error testing %v: expected %v but got %v", test.testName, test.expect, got)
		}
	}
}
