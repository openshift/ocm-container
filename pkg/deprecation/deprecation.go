package deprecation

import (
	"fmt"
	"strings"
)

func Message(deprecated string, alternative ...interface{}) {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("WARNING: %s is deprecated and will be removed in a future release", deprecated))
	if len(alternative) > 0 {
		s.WriteString(fmt.Sprintf("Please use %s instead", alternative[:]))
	}

	fmt.Println(s.String())
}
