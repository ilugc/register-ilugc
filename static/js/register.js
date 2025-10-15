let global = {
    register: document.getElementById("register"),
    status: document.getElementById("status"),
    gdiv: document.getElementById("gdiv"),
}

let redirectLink = function(chksum) {
    link ="/participant/" + chksum;
    location.assign(link);
}

let showMessage = function(message) {
    global.status.innerText = message;
}

fetch("/isclosed/", {
    method: "GET",
}).then((response) => {
    return response.text()
}, (err) => {
    showMessage(err.toString());
}).then((body) => {
    resp = JSON.parse(body);
    if (resp.isclosed === true) {
	global.gdiv.style.display = "none";
	showMessage("Registration Closed");
    }
});

global.register.addEventListener("focusout", (event) => {
    global.status.innerText = "";
});

global.register.addEventListener("click", (event) => {
    pname = document.getElementById("participant_name");
    if (pname.value.length < 0
	|| /^[A-Za-z ]+$/.test(pname.value) == false) {
	showMessage("invalid name");
	return;
    }

    pemail = document.getElementById("participant_email");
    if (pemail.value.length < 0	||
	/@/.test(pemail.value) == false) {
	showMessage("invalid email");
	return;
    }

    pmobile = document.getElementById("participant_mobile");
    if (pmobile.value.length < 10
	|| pmobile.value.length > 13
	|| /^[+0-9]+$/.test(pmobile.value) == false) {
	showMessage("invalid mobile number");
	return;
    }

    porg = document.getElementById("participant_org");
    if (porg.value.length < 0
	|| /^[A-Za-z0-9 ]+$/.test(porg.value) == false) {
	showMessage("invalid organization");
	return;
    }

    pplace = document.getElementById("participant_place");
    if (pplace.value.length < 0
	|| /^[A-Za-z ]+$/.test(pplace.value) == false) {
	showMessage("invalid place");
	return;
    }

    fetch("/register/", {
	method: "POST",
	header: {
	    "Content-Type": "application/json"
	},
	body: JSON.stringify({name: pname.value, email: pemail.value, mobile: pmobile.value, org: porg.value, place: pplace.value})
    }).then((response) => {
	return response.text();
    }, (err) => {
	showMessage(err.toString());
    }).then((body) => {
	try {
	    resp = JSON.parse(body);
	    redirectLink(resp.chksum);
	}
	catch(err) {
	    showMessage(body);
	}
    });
});
