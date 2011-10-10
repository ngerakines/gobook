
package main

import (
	"time"
	"strings"
	"strconv"
)

type Entry struct {
	Id, Message string
	When int64
}

type EntryGroup struct {
	Key string
	Entries []Entry
}

type EntryTag struct {
	Value string
	count int
}

func (entry Entry) PrettyDate() string {
	utc_time := time.SecondsToLocalTime(entry.When)
	value := utc_time.Format(time.RFC822)
	return value
}

func (entryGroup EntryGroup) PrettyDate() string {
	parts := strings.Split(entryGroup.Key, "-")
	when := time.LocalTime()
	if value, err := strconv.Atoi64(parts[0]); err == nil {
		when.Year = value
	}
	if value, err := strconv.Atoi(parts[1]); err == nil {
		when.Month = value
	}
	if value, err := strconv.Atoi(parts[2]); err == nil {
		when.Day = value
	}
	value := when.Format("Mon, 02 Jan 2006")
	return value
}

func (entryGroup EntryGroup) PrettyTimeOfDay() string {
	parts := strings.Split(entryGroup.Key, "-")
	if len(parts) != 4 {
		return "Unknown"
	}
	if value, err := strconv.Atoi(parts[3]); err == nil {
		switch value {
			case 0:
				return "Morning"
			case 1:
				return "Afternoon"
			case 2:
				return "Evening"
			default:
				return "Night"
		}
	}
	return "Unknown"
}

func (entry Entry) Tags() []EntryTag {
	tagStrings := getTags(entry.Id)
	tags := make([]EntryTag, len(tagStrings))
	for index, tag := range tagStrings {
		tags[index] = EntryTag{tag, 0}
	}
	return tags
}
