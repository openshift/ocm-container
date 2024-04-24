package deprecation

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Message takes a string representing the name of the deprecated item and an optional
// variadic of interfaces of alternative and returns a formatted string
func Message(deprecated string, alternative ...interface{}) string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("WARNING: '%s' is deprecated and will be removed in a future release.", deprecated))
	if len(alternative) > 0 {
		// Space before "Please" is intentional
		s.WriteString(fmt.Sprintf(" Please use '%s' instead", alternative[:]...))
	}

	return s.String()

}

// Message takes a string representing the name of the deprecated item and an optional
// variadic of interfaces of alternative and returns a short formatted string suitable
// for help and usage output
func ShortMessage(deprecated string, alternative ...interface{}) string {
	var s strings.Builder

	s.WriteString("(deprecated")
	if len(alternative) > 0 {
		s.WriteString(fmt.Sprintf(": use '%s' instead", alternative[:]...))
	}
	s.WriteString(")")

	return s.String()
}

// Print takes a string representing the name of the deprecated item and an optional
// variadic of interfaces of alternative and prints a formatted warning message to the console
func Print(deprecated string, alternative ...interface{}) {
	log.Warn(Message(deprecated, alternative))
}
