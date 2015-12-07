package gbytes_test

import (
	. "starwars-countdown/vendor/_nuts/github.com/onsi/ginkgo"
	. "starwars-countdown/vendor/_nuts/github.com/onsi/gomega"

	"testing"
)

func TestGbytes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gbytes Suite")
}
