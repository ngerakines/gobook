package main

import (
	"fmt"
	"time"
	"log"
	"sort"
	"strings"
)

func getTags(id string) []string {
	// NKG: Convert to prepared statement.
	err := db.Query("select tag from tags where id = \"" + id + "\"")
	if err != nil {
	    return []string{}
	}
	result, err := db.UseResult()
	if err != nil {
	    return []string{}
	}
	tags := make([]string, 100)
	tag_count := 0
	for {
	    row := result.FetchMap()
	    if row == nil {
	        break
	    }
		tag := string([]uint8( row["tag"].([]uint8)  ))
		tags[tag_count] = tag
		tag_count++
	}
	// NKG: Do I really have to fucking call this after every query?!
	db.FreeResult()

	return trimTagList(tags, tag_count)
}

func storeEntry(id UUID, message string, tags []string) {
	when := getTime(tags)
	stmt, err := db.Prepare("insert into entries values (?, ?, ?, ?)")
	if err != nil {
		return
	}
	if error := stmt.BindParams(id.String(), when, message, 0); error != nil {
		return
	}
	if error := stmt.Execute(); error != nil {
		return
	}
	for _, tag := range tags {
		storeTag(id, tag)
		storeReverseTag(id, tag)
	}
}

func getTime(tags []string) int64 {
	for _, tag := range tags {
		if strings.Index(tag, "!") == 0 {
			when := parseTime(tag)
			return when.Seconds()
		}
	}
	return time.Seconds()
}

func storeTag(entryId UUID, tag string) {
	stmt, err := db.Prepare("insert into tags values (?, ?)")
	if err != nil {
		return
	}
	if error := stmt.BindParams(entryId.String(), tag); error != nil {
		return
	}
	if error := stmt.Execute(); error != nil {
		return
	}
}

func storeReverseTag(entryId UUID, tag string) {
	stmt, err := db.Prepare("insert into tags_reverse values (?, ?)")
	if err != nil {
		return
	}
	if error := stmt.BindParams(tag, entryId.String()); error != nil {
		return
	}
	if error := stmt.Execute(); error != nil {
		return
	}
}

func getEntries() []Entry {
	err := db.Query("select * from entries limit 100")
	if err != nil {
	    return []Entry{}
	}
	result, err := db.UseResult()
	if err != nil {
	    return []Entry{}
	}
	entries := make([]Entry, 100)
	current := 0
	for {
	    row := result.FetchMap()
	    if row == nil {
	        break
	    }
		var entry Entry
		entry.Id = string([]uint8( row["id"].([]uint8)  ))
		entry.Message = string([]uint8( row["message"].([]uint8)  ))
		entry.When = row["date"].(int64)
		entries[current] = entry
		current++
	}
	// NKG: Do I really have to fucking call this after every query?!
	db.FreeResult()

	return trimEntryList(entries, current)
}

func trimEntryList(old_entries []Entry, size int) []Entry {
	entries := make([]Entry, size)
	for i := 0; i < size; i++ {
	        entries[i] = old_entries[i]
	}
	return entries
}

func groupEntries(entries []Entry) map[string][]Entry {
	groups := make(map[string][]Entry)
	// NKG: May seem strange, but I'm using another map to track the size
	// of the individual groups within the groups map. Will find a better way
	// to do this ...
	meta_group_entries := make(map[string]int)
	// NKG: Every time we place an entry the default group list size shrinks.
	count_down := len(entries)
	for _, entry := range entries {
		tod, utc_time := getTimeOfDay(entry.When)
		key := fmt.Sprintf("%d-%d-%d-%d", utc_time.Year, utc_time.Month, utc_time.Day, tod)
		if group_entry, ok := groups[key]; ok {
			index := meta_group_entries[key]
			group_entry[index] = entry
			index++
			meta_group_entries[key] = index
			groups[key] = group_entry
		} else {
			group_entry := make([]Entry, count_down)
			group_entry[0] = entry
			groups[key] = group_entry
			meta_group_entries[key] = 1
			count_down--
		}
	}
	return trimGroupedEntries(groups, meta_group_entries)
}

func getTimeOfDay(when int64) (tod int, utc_time *time.Time) {
	utc_time = time.SecondsToLocalTime(when)
	tod = 0 // default to morning (midnight to noon)
	switch {
		case utc_time.Hour < 4:
			tod = 3 // night, shift day back 1
		case utc_time.Hour < 12:
			tod = 0 // morning
		case utc_time.Hour < 17:
			tod = 1 // afternoon
		default:
			tod = 2 // evening
	}
	if tod == 3 {
		// Quick hack to mark night time as part of the previous day
		utc_time = time.SecondsToLocalTime(when - 86400)
	}
	return
}

func trimGroupedEntries(old_grouped_entries map[string][]Entry, meta_group_entries map[string]int) map[string][]Entry {
	grouped_entries := make(map[string][]Entry)
	for key, group_entries := range old_grouped_entries {
		size := meta_group_entries[key]
		grouped_entries[key] = reverseEntries(trimEntryList(group_entries, size))
	}
	return grouped_entries
}

func reverseEntries(old_entries []Entry) []Entry {
	entries := make([]Entry, len(old_entries))
	i := 0
	j := len(old_entries) - 1;
	for i < len(old_entries) {
		entries[j] = old_entries[i]
		i++
		j--
	}
	return entries
}

func groupedEntriesToEntryGroups(entries map[string][]Entry) []EntryGroup {
	keys := make([]string, len(entries))
	keyIndex := 0
	for key := range entries {
		keys[keyIndex] = key
		keyIndex++
	}
	sort.Strings(keys)

	entryGroupList := make([]EntryGroup, len(entries))
	for index, key := range keys {
		var entryGroup EntryGroup
		entryGroup.Key = key
		entryGroup.Entries = entries[key]
		entryGroupList[index] = entryGroup
		index++
	}
	return entryGroupList
}

func reverseEntryGroups(old_groups []EntryGroup) []EntryGroup {

	groups := make([]EntryGroup, len(old_groups))
	i := 0
	j := len(old_groups) - 1;
	for i < len(old_groups) {
		groups[j] = old_groups[i]
		i++
		j--
	}
	return groups
}

func dumpGroupedEntries(grouped_entries map[string][]Entry) {
	for key, group_entries := range grouped_entries {
		log.Println(key)
		for index, entry := range group_entries {
			log.Printf("#%d %s\n", index, entry.Id)
		}
	}
}

func trimTagList(old_tags []string, size int) []string {
	entries := make([]string, size)
	for i := 0; i < size; i++ {
	        entries[i] = old_tags[i]
	}
	return entries
}
