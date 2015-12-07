package format_test

import (
	. "starwars-countdown/vendor/_nuts/github.com/onsi/ginkgo"
	. "starwars-countdown/vendor/_nuts/github.com/onsi/gomega"

	"testing"
)

func TestFormat(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Format Suite")
}
