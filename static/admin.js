var base_string = "<br /><br /><input type=text name=\"Questions.N.Question\" placeholder=\"Question N Text\" /><br />\
			<input type=text name=\"Questions.N.Answers.0\" placeholder=\"Answer 1\" /><br />\
			<input type=text name=\"Questions.N.Answers.1\" placeholder=\"Answer 2\" /><br />\
			<input type=text name=\"Questions.N.Answers.2\" placeholder=\"Answer 3\" /><br />\
			<input type=text name=\"Questions.N.Answers.3\" placeholder=\"Answer 4\" /><br />\
			<label for=\"Questions.N.correct\">Which answer is correct? </label>\
			<select name=\"Questions.N.correct\"><option value=0>1</option><option value=1>2</option><option value=2>3</option><option value=3>4</option></select><br />";

var q_n = 0;

var do_formatting = function() {
	q_n += 1;
	var new_string = base_string;
	new_string = new_string.replace("N", q_n);
	new_string = new_string.replace("N", q_n + 1);
	for (var i = 0; i < 6; i++) {
		new_string = new_string.replace("N", q_n);
	}
	return new_string;
}

var addq = function() {
	document.getElementById("questions").innerHTML += do_formatting();
}
