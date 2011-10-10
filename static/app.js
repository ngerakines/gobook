
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

function renderTags(entry_div) {
	tags = Array()
	people = Array()
	other = Array()
	for (i = 1; i < arguments.length; i++) {
		if (arguments[i].substring(0, 1) == "#") {
			tags.push(arguments[i])
			allTags.push(arguments[i])
		} else if (arguments[i].substring(0, 1) == "@") {
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
		out += " <a href=\"#\">" + tags[i] + "</a>";
	}
	if (people.length > 0) {
		word = "with"
		if (out.length == 0) {
			word = "With"
		}
		out += " " + word + " ";
		for (i = 0; i < people.length; i++ ) {
			out += " <a href=\"#\">" + people[i] + "</a>";
		}
	}
	/* if (other.length == 1) {
		out += other[0]
	} */
	$("#" + entry_div).append("<p>" + out + "</p>");
	renderTagsList()
}

function renderTagsList() {
	$('#tagList').empty();
	allTags.sort();
	allTags = allTags.unique();
	allPeople.sort();
	allPeople = allPeople.unique();
	for (i = 0; i < allPeople.length; i++) {
		$("#tagList").append("<li>" + allPeople[i] + "</li>");
	}
	for (i = 0; i < allTags.length; i++) {
		$("#tagList").append("<li>" + allTags[i] + "</li>");
	}
}
