package main

import (
	"strings"
)

/*
splitTags("Nick Carolyn Vanessa Hannah")
splitTags("\"Hello World\"")
splitTags("\"@Carolyn Gerakines\" #dinner #date")
splitTags("#meeting \"@Steve McGarrity\" #port #battle.net    \"\"")
*/
func splitTags(value string) []string {
	reader := strings.NewReader(value)
	// being greedy with allocating tags array size
	tags := make([]string, len(value))
	tag_count := 0
	buffer := ""
	isInQuote := false
	lastRune := 0
	appendBuffer := false
	for {
		appendBuffer = true
		rune, _, err := reader.ReadRune()
		if err != nil {
			if len(buffer) > 0 {
				tags[tag_count] = buffer
				tag_count++
				buffer = ""
			}
			break
		}
		if rune == 32 {
			if isInQuote == false {
				appendBuffer = false
				if len(buffer) > 0 {
					tags[tag_count] = buffer
					tag_count++
					buffer = ""
				}
			}
		}
		if rune == 34 {
			if lastRune == 32 || lastRune == 34 || lastRune == 0 {
				appendBuffer = false
				isInQuote = true
			}
			if len(buffer) > 0 {
				isInQuote = false
				appendBuffer = false
				tags[tag_count] = buffer
				tag_count++
				buffer = ""
			}
		}
		if appendBuffer {
			buffer = buffer + string(rune)
		}
		lastRune = rune
	}
	return trimTagList(tags, tag_count)
}
