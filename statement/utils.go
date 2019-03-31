package statement

import (
	"strconv"
	"strings"
)

func placeHolder(start int, count int) string {
	var b strings.Builder
	for i := 0; i < count; i++ {
		b.WriteString("$")
		b.WriteString(strconv.Itoa(start + i))
		if i < count-1 {
			b.WriteString(", ")
		}
	}
	return b.String()
}
