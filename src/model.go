
package main

import (
	"time"
	"bytes"
	"strings"
	"strconv"
	"github.com/russross/blackfriday"
)

type Entry struct {
	Id, Message string
	When int64
}

type EntryGroup struct {
	Key string
	Entries []*Entry
}

type EntryTag struct {
	Value string
	count int
}

func (entryGroup *EntryGroup) AddEntry(entry *Entry) {
	if entryGroup.Entries == nil {
		entryGroup.Entries = make([]*Entry, 0, 100)
	}
	n := len(entryGroup.Entries)
	if n + 1 > cap(entryGroup.Entries) {
		s := make([]*Entry, n, 2 * n + 1)
		copy(s, entryGroup.Entries)
		entryGroup.Entries = s
	}
	entryGroup.Entries = entryGroup.Entries[0 : n + 1]
	entryGroup.Entries[n] = entry
}

func (entry Entry) PrettyDate() string {
	utc_time := time.SecondsToLocalTime(entry.When)
	value := utc_time.Format(time.RFC822)
	return value
}

func (entry Entry) PrettyMessage() string {
	buffer := bytes.NewBufferString(entry.Message)
	output := blackfriday.MarkdownCommon(buffer.Bytes())
	return string(output)
}

func (entryGroup EntryGroup) PrettyGroupDate() string {
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
	retval := when.Format("_2 Jan 2006")
	return retval
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
