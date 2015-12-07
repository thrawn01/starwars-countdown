package table_test

import (
	"strings"

	. "starwars-countdown/vendor/_nuts/github.com/onsi/ginkgo/extensions/table"

	. "starwars-countdown/vendor/_nuts/github.com/onsi/ginkgo"
	. "starwars-countdown/vendor/_nuts/github.com/onsi/gomega"
)

var _ = Describe("Table", func() {
	DescribeTable("a simple table",
		func(x int, y int, expected bool) {
			Ω(x > y).Should(Equal(expected))
		},
		Entry("x > y", 1, 0, true),
		Entry("x == y", 0, 0, false),
		Entry("x < y", 0, 1, false),
	)

	type ComplicatedThings struct {
		Superstructure string
		Substructure   string
		Count          int
	}

	DescribeTable("a more complicated table",
		func(c ComplicatedThings) {
			Ω(strings.Count(c.Superstructure, c.Substructure)).Should(BeNumerically("==", c.Count))
		},
		Entry("with no matching substructures", ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "emir",
			Count:          0,
		}),
		Entry("with one matching substructure", ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "sheep",
			Count:          1,
		}),
		Entry("with many matching substructures", ComplicatedThings{
			Superstructure: "the sixth sheikh's sixth sheep's sick",
			Substructure:   "si",
			Count:          3,
		}),
	)

	PDescribeTable("a failure",
		func(value bool) {
			Ω(value).Should(BeFalse())
		},
		Entry("when true", true),
		Entry("when false", false),
		Entry("when malformed", 2),
	)
})
