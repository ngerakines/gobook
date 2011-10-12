package main

import (
	"strings"
	"regexp"
	"log"
	"time"
	"strconv"
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

func clampRange(size, min, max int) int {
	if size < min {
		return size
	}
	if size > max {
		return max
	}
	return size
}

/*
parseTime("3:45p")
parseTime("9:45a")
parseTime("2011-10-07")
parseTime("2011-10-06 10:45p")
*/
func parseTime(val string) *time.Time {
	val = strings.ToLower(val)
	when := time.LocalTime()

	timeA, err := regexp.Compile("([0-9]+):([0-9]+)(a|p)?")
	if err != nil {
		log.Println(err)
		return when
	}

	timeB, err := regexp.Compile("([0-9]+)\\-([0-9]+)\\-([0-9]+)")
	if err != nil {
		log.Println(err)
		return when
	}

	if match := timeA.MatchString(val); match {
		matches := timeA.FindAllStringSubmatch(val, -1)
		submatches := matches[0]
		if hour, err := strconv.Atoi(submatches[1]); err == nil {
			when.Hour = hour
		}
		if minute, err := strconv.Atoi(submatches[2]); err == nil {
			when.Minute = minute
		}
		if submatches[3] != "" {
			ap := submatches[3]
			if ap == "p" {
				hour := when.Hour
				hour += 12
				when.Hour = hour
			}
		}
	}
	if match := timeB.MatchString(val); match {
		matches := timeB.FindAllStringSubmatch(val, -1)
		submatches := matches[0]
		if year, err := strconv.Atoi64(submatches[1]); err == nil {
			when.Year = year
		}
		if month, err := strconv.Atoi(submatches[2]); err == nil {
			when.Month = month
		}
		if day, err := strconv.Atoi(submatches[3]); err == nil {
			when.Day = day
		}
	}
	return when
}
