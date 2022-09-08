package cli

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"testing"
)

func TestReadAnswerInvalidDst(t *testing.T) {
	p := &survey.Input{
		Message: "demo",
	}
	var dst string
	err := readAnswer(p, dst)
	if !errors.Is(err, errPointerExpected) {
		t.Fatal(err)
	}
}
