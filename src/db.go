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
	result, err := db.StoreResult()
	if err != nil {
	    return []string{}
	}
	tags := make([]string, result.RowCount())
	for index, row := range result.FetchRows() {
		tag := string([]uint8( row[0].([]uint8)  ))
		tags[index] = tag
	}
	db.FreeResult()
	return tags
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

func updateTag(oldName, newName string) {
	stmt, err := db.Prepare("update tags set tag = ? where tag = ?")
	if err != nil {
		return
	}
	if error := stmt.BindParams(newName, oldName); error != nil {
		return
	}
	if error := stmt.Execute(); error != nil {
		return
	}
}

func getEntriesFromTag(tag string) []*Entry {
	query := fmt.Sprintf("select id, message, date from entries where id in (select id from tags where tag = \"%s\") order by date DESC", db.Escape(tag) )
	err := db.Query(query)
	if err != nil {
	    return []*Entry{}
	}
	result, err := db.StoreResult()
	if err != nil {
	    return []*Entry{}
	}
	entries := make([]*Entry, result.RowCount())
	for index, row :=  range result.FetchRows() {
		entry := new(Entry)
		entry.Id = string([]uint8( row[0].([]uint8)  ))
		entry.Message = string([]uint8( row[1].([]uint8)  ))
		entry.When = row[2].(int64)
		entries[index] = entry
	}
	// NKG: Do I really have to fucking call this after every query?!
	db.FreeResult()

	return entries
}

func getEntry(id string) *Entry {
	query := fmt.Sprintf("select id, message, date from entries where id = \"%s\"", db.Escape(id))
	err := db.Query(query)
	if err != nil {
	    return nil
	}
	result, err := db.StoreResult()
	if err != nil {
	    return nil
	}
	if result.RowCount() != 1 {
		return nil
	}
	entry := new(Entry)
	for _, row :=  range result.FetchRows() {
		entry.Id = string([]uint8( row[0].([]uint8)  ))
		entry.Message = string([]uint8( row[1].([]uint8)  ))
		entry.When = row[2].(int64)
	}
	// NKG: Do I really have to fucking call this after every query?!
	db.FreeResult()

	return entry
}

func getEntries() []*Entry {
	err := db.Query("select id, message, date from entries order by date DESC limit 250")
	if err != nil {
	    return []*Entry{}
	}
	result, err := db.StoreResult()
	if err != nil {
	    return []*Entry{}
	}
	entries := make([]*Entry, result.RowCount())
	for index, row :=  range result.FetchRows() {
		entry := new(Entry)
		entry.Id = string([]uint8( row[0].([]uint8)  ))
		entry.Message = string([]uint8( row[1].([]uint8)  ))
		entry.When = row[2].(int64)
		entries[index] = entry
	}
	// NKG: Do I really have to fucking call this after every query?!
	db.FreeResult()

	return entries
}

func groupEntries(entries []*Entry) map[string]*EntryGroup {
	groups := make(map[string]*EntryGroup)
	// NKG: May seem strange, but I'm using another map to track the size
	// of the individual groups within the groups map. Will find a better way
	// to do this ...
	// meta_group_entries := make(map[string]int)
	// NKG: Every time we place an entry the default group list size shrinks.
	// count_down := len(entries)
	for _, entry := range entries {
		tod, utc_time := getTimeOfDay(entry.When)
		key := fmt.Sprintf("")                  
                if(utc_time.Day < 10){
			key = fmt.Sprintf("%d-%d-0%d-%d", utc_time.Year, utc_time.Month, utc_time.Day, tod)
		}else {
			key = fmt.Sprintf("%d-%d-%d-%d", utc_time.Year, utc_time.Month, utc_time.Day, tod)
		}
		entryGroup, ok := groups[key]
		if ok == false {
			entryGroup = new(EntryGroup)
			entryGroup.Key = key
			groups[key] = entryGroup
		}
		entryGroup.AddEntry(entry)
		/* if group_entry, ok := groups[key]; ok {
			index := meta_group_entries[key]
			group_entry[index] = entry
			index++
			meta_group_entries[key] = index
			groups[key] = group_entry
		} else {
			group_entry := make([]Entry, 1, count_down)
			group_entry[0] = entry
			groups[key] = group_entry
			meta_group_entries[key] = 1
			count_down--
		} */
	}
	return groups
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

func groupedEntriesToEntryGroups(entries map[string][]*Entry) []EntryGroup {
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

func flattenEntryGroups(entryGroups map[string]*EntryGroup) []*EntryGroup {
	keys := make([]string, len(entryGroups))
	keyIndex := 0
	for key, _ := range entryGroups {
		keys[keyIndex] = key
		keyIndex++
	}
	sort.Strings(keys)

	groups := make([]*EntryGroup, len(keys))
	j := len(keys) - 1
	for _, key := range keys {
		groups[j] = entryGroups[key]
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

