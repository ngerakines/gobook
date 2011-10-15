
Array.prototype.unique =
  function() {
    var a = [];
    var l = this.length;
    for(var i=0; i<l; i++) {
      for(var j=i+1; j<l; j++) {
        // If this[i] is found later in the array
        if (this[i] === this[j])
          j = ++i;
      }
      a.push(this[i]);
    }
    return a;
  };

var allTags = Array();
var allPeople = Array();
var groups = { };

function renderTags(entry_div) {
	tags = Array()
	people = Array()
	other = Array()

	var $parent = $("#" + entry_div).parent()
	var $group_id = $parent.attr("id");
	if (!($group_id in groups)) {
		groups[$group_id] = { 'tags': [], 'people': [] };
	}

	for (i = 1; i < arguments.length; i++) {
		if (arguments[i].substring(0, 1) == "#") {
			groups[$group_id]['tags'].push(arguments[i])
			tags.push(arguments[i])
			allTags.push(arguments[i])
		} else if (arguments[i].substring(0, 1) == "@") {
			groups[$group_id]['people'].push(arguments[i])
			people.push(arguments[i])
			allPeople.push(arguments[i])
		} else {
			other.push(arguments[i])
		}
	}
	tags.sort();
	people.sort();
	var out = "";
	for (i = 0; i < tags.length; i++ ) {
		out += " <a class=\"tag\" href=\"#\">" + tags[i] + "</a>";
	}
	if (people.length > 0) {
		word = "with"
		if (out.length == 0) {
			word = "With"
		}
		out += " " + word + " ";
		for (i = 0; i < people.length; i++ ) {
			out += " <a class=\"person\" href=\"#\">" + people[i] + "</a>";
		}
	}
	/* if (other.length == 1) {
		out += other[0]
	} */
	$("#" + entry_div).append("<p>" + out + "</p>");
}

function renderTagsList() {
	$('#tagList').empty();
	allTags.sort();
	allTags = allTags.unique();
	allPeople.sort();
	allPeople = allPeople.unique();
	for (i = 0; i < allPeople.length; i++) {
		var $new_person = $('<li class="person">' + allPeople[i] + '</li>');
		$('#tagList').append($new_person)
	}
	for (i = 0; i < allTags.length; i++) {
		var $new_tag = $('<li class="tag">' + allTags[i] + '</li>');
		$('#tagList').append($new_tag)
	}
}
