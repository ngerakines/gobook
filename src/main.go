package main

import (
	"os"
	"io"
	"fmt"
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
	"github.com/Philio/GoMySQL"
	"github.com/garyburd/twister/server"
	"github.com/garyburd/twister/web"
	// "strings"
	"log"
	// "time"
	"url"
)

var (
	db *mysql.Client
	db_err os.Error
	session *mgo.Session
	session_err os.Error
)

func displayIndex(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/index.html", map[string]string{"c":"world"}))
}

func createEntry(req *web.Request) {
	message := req.Param.Get("message")
	extra := req.Param.Get("extra")
	tags := splitTags(extra)
	when := getTime(tags)
	createItem(message, extra, tags, when)
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/index.html", map[string]string{"c":"world"}))
}

func displayArchive(req *web.Request) {
	result := getItems(nil, 250)
	groupedItems := groupItems(result)
	flatItemGroups := flattenItemGroups(groupedItems)

	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	params := make(map[string]interface{})
	params["entry_groups"] = flatItemGroups
	io.WriteString(w, RenderFile("templates/archive.html", params))
}

func renameTag(req *web.Request) {
	// oldTag := req.Param.Get("oldTag")
	newTag := req.Param.Get("newTag")
	// updateTag(oldTag, newTag)
	url := fmt.Sprintf("/tag/%s", url.QueryEscape(newTag))
	req.Redirect(url, false)
}

func displayTag(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	params := make(map[string]interface{})
	if tag, ok := req.URLParam["tag"]; ok {
		result := getItems(bson.M{"tags": bson.M{"$in": []string{tag}}}, 1000)
		groupedItems := groupItems(result)
		params["tag"] = tag
		params["entry_groups"] = flattenItemGroups(groupedItems)

	}
	io.WriteString(w, RenderFile("templates/tag.html", params))
}

func getItems(query bson.M, limit int) []*Item {
	c := session.DB("gobook").C("items")
	var result []*Item
	iter := c.Find(query).Sort(bson.M{ "when": -1}).Limit(limit).Iter()
	c_err := iter.All(&result)

	if c_err != nil {
		log.Println(c_err)
	}
	return result
}

/* func displayEntry(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
        params := make(map[string]interface{})
        if id, ok := req.URLParam["id"]; ok {
		log.Println(id)
                // entry := getEntry(id)
                // params["entry"] = entry
        }
        io.WriteString(w, RenderFile("templates/entry.html", params))
} */

/* func displayMonth(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/month.html", map[string]string{}))
} */

func migrate(req *web.Request) {

	err := db.Query("select id, message, date from entries")
	if err != nil {
		panic(err)
	}
	result, err := db.StoreResult()
	if err != nil {
		panic(err)
	}
	itemIndex := make(map[string]Item, result.RowCount())
	for _, entry_row := range result.FetchRows() {
		id := string( []uint8( entry_row[0].([]uint8) ) )
		message := string([]uint8(entry_row[1].([]uint8)))
		when := entry_row[2].(int64)
		itemIndex[id] = Item{message, "", make([]string, 0), bson.MongoTimestamp(when) }
	}
	db.FreeResult()

	for id, item := range itemIndex {
		tags := getTags(id)
		// log.Println(id, tags)
		// item.Tags = tags
		// itemIndex[id] = item
		createItem(item.Message, "", tags, int64(item.When))
	}

	log.Println(itemIndex)

	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, "OK")

}

func getTags(id string) []string {
	// NKG: Convert to prepared statement.
	err := db.Query("select tag from tags where id = \"" + id + "\"")
	if err != nil {
		panic(err)
	}
	result, err := db.StoreResult()
	if err != nil {
		panic(err)
	}
	tags := make([]string, result.RowCount())
	for index, row := range result.FetchRows() {
		tag := string([]uint8( row[0].([]uint8)  ))
		tags[index] = tag
	}
	db.FreeResult()
	return tags
}


func main() {
	/*
	log.Println(splitTags("Nick Carolyn Vanessa Hannah"))
	log.Println(splitTags("\"Hello World\""))
	log.Println(splitTags("\"@Carolyn Gerakines\" #dinner #date"))
	log.Println(splitTags("#meeting \"@Steve McGarrity\" #port #battle.net    \"\""))
	log.Println(splitTags("#api-wow +3h"))
	*/

	session, session_err = mgo.Mongo("localhost")
	if session_err != nil {
		panic(session_err)
	}
	defer session.Close()

	db, db_err = mysql.DialTCP("localhost", "root", "asd123", "gobook")
	if db_err != nil {
		log.Println(db_err)
	    os.Exit(1)
	}

	port := ":8080"
	if envPort := os.Getenv("GOBOOK_PORT"); envPort != "" {
		port = envPort
	}

	h := web.FormHandler(10000, false,
		web.NewRouter().
			Register("/", "GET", displayIndex, "POST", createEntry).
			Register("/view/<entry:.*>", "GET", displayEntry).
			Register("/thread/<id:.*>", "GET", displayThread).
			Register("/calendar", "GET", displayCalendar).
			Register("/calendar/<year:.*>/<month:.*>/<day:.*>", "GET", displayDay).
			Register("/calendar/<year:.*>/<month:.*>", "GET", displayMonth).
			Register("/calendar/<year:.*>", "GET", displayYear).

			// Register("/migrate", "GET", migrate).
			Register("/archive", "GET", displayArchive).
			Register("/tag/<tag:.*>", "GET", displayTag).
			// Register("/entry/<id:.*>", "GET", displayEntry).
			Register("/api/tag/rename/", "POST", renameTag).
			Register("/summary/<year:.*>/<month:.*>", "GET", displayMonth).
			Register("/static/<path:.*>", "GET", web.DirectoryHandler("./static/", new(web.ServeFileOptions))))
	server.Run(port, h)
}

func displayEntry(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/entry.html", map[string]string{}))
}

func displayThread(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/thread.html", map[string]string{}))
}

func displayCalendar(req *web.Request) {

	job := mgo.MapReduce{
		Map:      "function() { day = new Date(this.when.i * 1000); emit(day.getFullYear().toString() + '-' + day.getMonth().toString() + '-' + day.getDate().toString(), 1); }",
		// Reduce:   "function(key, values) { var count = 0; values.forEach(function(v) { count += v['count']; }); return count; }",
		Reduce:   "function(key, values) { return Array.sum(values) }",
	}
	var result []struct { Id string "_id"; Count int "value" }
	collection := session.DB("gobook").C("items")
	_, err := collection.Find(nil).MapReduce(job, &result)

	if err != nil {
		log.Println(err)
	}
	log.Println(result);

	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/calendar.html", map[string]string{}))
}

func displayYear(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/year.html", map[string]string{}))
}

func displayMonth(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/month.html", map[string]string{}))
}

func displayDay(req *web.Request) {
	w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=\"utf-8\"")
	io.WriteString(w, RenderFile("templates/day.html", map[string]string{}))
}

