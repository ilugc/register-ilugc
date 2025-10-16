let global = {
    change: document.getElementById("change"),
    status: document.getElementById("status"),
    gdiv: document.getElementById("gdiv")
}

let showMessage = function(message) {
    global.status.innerText = message;
}

global.change.addEventListener("click", (event) => {
    body = {}
    formelements = document.getElementsByClassName("form");
    for (el of formelements) {
	if (el.value.length <= 0) {
	    continue
	}

	switch (el.id) {
	case "AdminPassword": {
	    body.AdminPassword = btoa(el.value);
	    break;
	}
	case "DefaultMax": {
	    body.DefaultMax = parseInt(el.value);
	    break;
	}
	case "StopRegistration": {
	    body.StopRegistration = el.value === "true" ? true : false;
	    break;
	}
	default: {
	    body[el.id] = el.value;
	}
	}
    }

    fetch("/config/", {
	method: "POST",
	header: {
	    "Content-Type": "application/json"
	},
	body: JSON.stringify(body)
    }).then((response) => {
	if (response.ok === true) {
	    global.gdiv.style.display = "none";
	}
	return response.text();
    }, (err) => {
	showMessage(err.toString());
    }).then((body) => {
	showMessage(body);
    });
});

global.change.addEventListener("focusout", (event) => {
    if (global.change.checkVisibility()) {
	global.status.innerText = "";
    }
});
