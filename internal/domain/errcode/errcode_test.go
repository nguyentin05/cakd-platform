package errcode_test

import (
	"errors"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
)

func TestDomainErrorWrapAndCheck(t *testing.T) {
	cause := errors.New("missing metadata.name")
	err := errcode.Wrap(errcode.CodeValidation, "invalid project spec", cause)

	if !errcode.IsCode(err, errcode.CodeValidation) {
		t.Fatalf("expected CodeValidation, got %v", err)
	}
	if errcode.IsCode(err, errcode.CodeAuth) {
		t.Fatalf("did not expect CodeAuth")
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected errors.Is to match wrapped cause")
	}
	if got := errcode.CodeOf(err); got != errcode.CodeValidation {
		t.Fatalf("expected CodeOf=%s, got %s", errcode.CodeValidation, got)
	}
}
