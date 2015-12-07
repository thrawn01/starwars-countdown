package main

import (
	"starwars-countdown/vendor/_nuts/github.com/Sirupsen/logrus"
	"starwars-countdown/vendor/_nuts/gopkg.in/gemnasium/logrus-airbrake-hook.v2"
)

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.TextFormatter) // default
	log.Hooks.Add(airbrake.NewHook(123, "xyz", "development"))
}

func main() {
	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	log.WithFields(logrus.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")

	log.WithFields(logrus.Fields{
		"omg":    true,
		"number": 100,
	}).Fatal("The ice breaks!")
}
