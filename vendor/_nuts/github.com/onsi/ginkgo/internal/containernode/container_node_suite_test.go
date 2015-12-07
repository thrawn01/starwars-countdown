package containernode_test

import (
	. "starwars-countdown/vendor/_nuts/github.com/onsi/ginkgo"
	. "starwars-countdown/vendor/_nuts/github.com/onsi/gomega"

	"testing"
)

func TestContainernode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Containernode Suite")
}
