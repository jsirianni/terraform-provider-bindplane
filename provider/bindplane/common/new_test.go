package common

import (
	"os"
	"testing"
)

const fakeValidUUID = "abcdefAB-0123-4ABC-ab12-CDEF01234567"

func TestNew(t *testing.T) {
	os.Setenv("BINDPLANE_API_KEY", fakeValidUUID)
	_, err := New()
	if err == nil {
		t.Errorf("Expected New to return a test connection error when using a fake and invalid api key")
	}
}

func TestCheckEnvEmpty(t *testing.T) {
	os.Setenv("BINDPLANE_API_KEY", "")
	if checkEnv() == nil {
		t.Errorf("Expected CheckEnv to return an error when BINDPLANE_API_KEY is empty")
	}
}

func TestCheckEnvUUID(t *testing.T) {
	os.Setenv("BINDPLANE_API_KEY", fakeValidUUID)
	if err := checkEnv(); err != nil {
		t.Errorf("Expected CheckEnv to return nil when using a valid uuid, got: " + err.Error())
	}

	os.Setenv("BINDPLANE_API_KEY", "abc")
	if checkEnv() == nil {
		t.Errorf("Expected CheckEnv to return an error when using an invalid uuid, got nil")
	}
}
