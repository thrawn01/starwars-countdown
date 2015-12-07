package testsuite_test

import (
	. "starwars-countdown/vendor/_nuts/github.com/onsi/ginkgo"
	. "starwars-countdown/vendor/_nuts/github.com/onsi/gomega"

	"testing"
)

func TestTestsuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testsuite Suite")
}
