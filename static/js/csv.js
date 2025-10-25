let global = {
    download: document.getElementById("download"),
    status: document.getElementById("status"),
    gdiv: document.getElementById("gdiv"),
    fromtime: document.getElementById("fromtime")
}

let redirectLink = function(hash) {
    link ="/csv/" + hash;
    location.assign(link);
}

let showMessage = function(message) {
    global.status.innerText = message;
}

global.download.addEventListener("click", (event) => {
    fromtime = global.fromtime.value
    body = {fromtime: fromtime}
    fetch("/csv/", {
	method: "POST",
	headers: {
	    "Content-Type": "application/json",
	    "Authorization": "Basic " + genAuth()
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
