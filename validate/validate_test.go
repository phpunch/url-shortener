package validate

import (
	"testing"
)

func TestCheckBlackList_Success(t *testing.T) {
	if err := CheckBlackList("http://www.facebook.com/321313"); err != nil {
		t.Fatalf("it should not be banned, err: %v", err)
	}
}
func TestCheckBlackList_Failed(t *testing.T) {
	if err := CheckBlackList("http://www.google.com/321313"); err == nil {
		t.Fatalf("it should be banned")
	}
}
