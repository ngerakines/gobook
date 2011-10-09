package main

import (
	"fmt"
	// "github.com/Philio/GoMySQL"
	"time"
	"log"
)


type ListOEntries []Entry
type GroupedEntries map[string]ListOEntries

type Entry struct {
	id, message string
	when int64
}

func storeEntry(id UUID, message string, tags []string) {
	fmt.Println(message)
	fmt.Println(tags)
	stmt, err := db.Prepare("insert into entries values (?, ?, ?, ?)")
	if err != nil {
		log.Println(err)
		return
	}
	if error := stmt.BindParams(id.String(), time.Seconds(), message, 0); error != nil {
		log.Println(error)
		return
	}
	if error := stmt.Execute(); error != nil {
		log.Println(error)
		return
	}
}

func getEntries() []Entry {
	err := db.Query("select * from entries limit 100")
	if err != nil {
		log.Println(err)
	    return []Entry{}
	}
	result, err := db.UseResult()
	if err != nil {
		log.Println(err)
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
		entry.id = string([]uint8( row["id"].([]uint8)  ))
		entry.message = string([]uint8( row["message"].([]uint8)  ))
		entry.when = row["date"].(int64)
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
		tod, utc_time := getTimeOfDay(entry.when)
		key := fmt.Sprintf("%d-%d-%d-%d", utc_time.Year, utc_time.Month, utc_time.Day, tod)
		fmt.Println(key)
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
	fmt.Println(utc_time)
	tod = 0 // default to morning (midnight to noon)
	switch {
		case utc_time.Hour < 4:
			tod = 3 // night, shift day back 1
		case utc_time.Hour < 12:
			tod = 0 // morning
		case utc_time.Hour < 5:
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
		grouped_entries[key] = trimEntryList(group_entries, size)
	}
	return grouped_entries
}

func dumpGroupedEntries(grouped_entries map[string][]Entry) {
	for key, group_entries := range grouped_entries {
		fmt.Println(key)
		for index, entry := range group_entries {
			fmt.Printf("#%d %s\n", index, entry.id)
		}
	}
}
