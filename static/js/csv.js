let global = {
    download: document.getElementById("download"),
    status: document.getElementById("status"),
    gdiv: document.getElementById("gdiv")
}

let redirectLink = function(hash) {
    link ="/csv/" + hash;
    location.assign(link);
}

let showMessage = function(message) {
    global.status.innerText = message;
}

global.download.addEventListener("click", (event) => {
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
	default: {
	    body[el.id] = el.value;
	}
	}
    }

    fetch("/csv/", {
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
	try {
	    resp = JSON.parse(body);
	    redirectLink(resp.hash);
	}
	catch(err) {
	    showMessage(body);
	}
    });
});

global.download.addEventListener("focusout", (event) => {
    if (global.download.checkVisibility()) {
	global.status.innerText = "";
    }
});
