package deprecation

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Message takes a string representing the name of the deprecated item and an optional
// variadic of interfaces of alternative and returns a formatted string
func Message(deprecated string, alternative ...interface{}) string {
	var s strings.Builder

	fmt.Fprintf(&s, "WARNING: '%s' is deprecated and will be removed in a future release.", deprecated)
	if len(alternative) > 0 {
		fmt.Fprintf(&s, " Please use '%s' instead", alternative[:]...)
	}

	return s.String()

}

// ShortMessage takes a string representing the name of the deprecated item and an optional
// variadic of interfaces of alternative and returns a short formatted string suitable
// for help and usage output
func ShortMessage(deprecated string, alternative ...interface{}) string {
	var s strings.Builder

	s.WriteString("(deprecated")
	if len(alternative) > 0 {
		fmt.Fprintf(&s, ": use '%s' instead", alternative[:]...)
	}
	s.WriteString(")")

	return s.String()
}

// Print takes a string representing the name of the deprecated item and an optional
// variadic of interfaces of alternative and prints a formatted warning message to the console
func Print(deprecated string, alternative ...interface{}) {
	log.Warn(Message(deprecated, alternative))
}
