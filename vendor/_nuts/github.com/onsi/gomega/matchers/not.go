package matchers

import (
	"starwars-countdown/vendor/_nuts/github.com/onsi/gomega/internal/oraclematcher"
	"starwars-countdown/vendor/_nuts/github.com/onsi/gomega/types"
)

type NotMatcher struct {
	Matcher types.GomegaMatcher
}

func (m *NotMatcher) Match(actual interface{}) (bool, error) {
	success, err := m.Matcher.Match(actual)
	if err != nil {
		return false, err
	}
	return !success, nil
}

func (m *NotMatcher) FailureMessage(actual interface{}) (message string) {
	return m.Matcher.NegatedFailureMessage(actual) // works beautifully
}

func (m *NotMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return m.Matcher.FailureMessage(actual) // works beautifully
}

func (m *NotMatcher) MatchMayChangeInTheFuture(actual interface{}) bool {
	return oraclematcher.MatchMayChangeInTheFuture(m.Matcher, actual) // just return m.Matcher's value
}
