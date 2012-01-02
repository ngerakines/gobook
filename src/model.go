
package main

import (
	"fmt"
	"time"
	"bytes"
	"sort"
	"strings"
	"strconv"
	"launchpad.net/gobson/bson"
	// "launchpad.net/mgo"
	"github.com/russross/blackfriday"
)

type Item struct {
	Message, Extra string
	Tags []string
	When bson.MongoTimestamp
}

type Note struct {
	Message, Extra string
	Parent bson.ObjectId
	When bson.Timestamp
}

func (item *Item) String() string {
	return fmt.Sprintf("%s | %s | %s | %b", item.Message, item.Extra, item.Tags, int64(item.When))
}

func createItem(Message, Extra string, Tags []string, When int64) {
	c := session.DB("gobook").C("items")
	c_err := c.Insert(&Item{Message, Extra, Tags, bson.MongoTimestamp(When)})
	if c_err != nil {
		panic(c_err)
	}
}

type ItemGroup struct {
	Key string
	Items []*Item
}

func (itemGroup *ItemGroup) String() string {
	return fmt.Sprintf("%s | %b | %s", itemGroup.Key, len(itemGroup.Items), itemGroup.Items)
}

func (itemGroup *ItemGroup) AddItem(item *Item) {
	if itemGroup.Items == nil {
		itemGroup.Items = make([]*Item, 0, 100)
	}
	n := len(itemGroup.Items)
	if n + 1 > cap(itemGroup.Items) {
		s := make([]*Item, n, 2 * n + 1)
		copy(s, itemGroup.Items)
		itemGroup.Items = s
	}
	itemGroup.Items = itemGroup.Items[0 : n + 1]
	itemGroup.Items[n] = item
}

func groupItems(items []*Item) map[string]*ItemGroup {
	groups := make(map[string]*ItemGroup)
	for _, item := range items {
		tod, utc_time := getTimeOfDay(int64(item.When))
		key := fmt.Sprintf("")
		if utc_time.Day < 10 {
			key = fmt.Sprintf("%d-%d-0%d-%d", utc_time.Year, utc_time.Month, utc_time.Day, tod)
		} else {
			key = fmt.Sprintf("%d-%d-%d-%d", utc_time.Year, utc_time.Month, utc_time.Day, tod)
		}
		itemGroup, ok := groups[key]
		if ok == false {
			itemGroup = new(ItemGroup)
			itemGroup.Key = key
			groups[key] = itemGroup
		}
		itemGroup.AddItem(item)
	}
	return groups
}

func flattenItemGroups(itemGroups map[string]*ItemGroup) []*ItemGroup {
	keys := make([]string, len(itemGroups))
	keyIndex := 0
	for key, _ := range itemGroups {
		keys[keyIndex] = key
		keyIndex++
	}
	sort.Strings(keys)

	groups := make([]*ItemGroup, len(itemGroups))
	j := len(keys) - 1
	for _, key := range keys {
		groups[j] = itemGroups[key]
		j--
	}
	return groups
}

type ItemTag struct {
	Value string
	count int
}

func (item Item) PrettyDate() string {
	utc_time := time.SecondsToLocalTime(int64(item.When))
	value := utc_time.Format(time.RFC822)
	return value
}

func (item Item) PrettyMessage() string {
	buffer := bytes.NewBufferString(item.Message)
	output := blackfriday.MarkdownCommon(buffer.Bytes())
	return string(output)
}

func (itemGroup ItemGroup) PrettyGroupDate() string {
	parts := strings.Split(itemGroup.Key, "-")
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

func (itemGroup ItemGroup) PrettyTimeOfDay() string {
	parts := strings.Split(itemGroup.Key, "-")
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

func (item Item) ItemTags() []ItemTag {
	tagStrings := item.Tags
	tags := make([]ItemTag, len(tagStrings))
	for index, tag := range tagStrings {
		tags[index] = ItemTag{tag, 0}
	}
	return tags
}

type ItemGroupWrapper struct {
	Key string
	Id int
	Item *Item
}

func (itemGroup *ItemGroup) GetItems() []ItemGroupWrapper {
	itemGroupWrappers := make([]ItemGroupWrapper, len(itemGroup.Items))
	for index, item := range itemGroup.Items {
		itemGroupWrappers[index] = ItemGroupWrapper{itemGroup.Key, index, item}
	}
	return itemGroupWrappers
}

func (itemGroupWrapper ItemGroupWrapper) ItemId() string {
	return fmt.Sprintf("%s-%b", itemGroupWrapper.Key, itemGroupWrapper.Id)
}

