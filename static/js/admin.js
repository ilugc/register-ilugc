let encodePassword = function(password) {
    if (password !== undefined && password.length <= 0) {
	console.log("empty or invalid password")
	return "";
    }
    utf8encoder = new TextEncoder();
    utf8bytes = utf8encoder.encode(password);
    randarray = new Uint8Array(utf8bytes.length);
    crypto.getRandomValues(randarray);
    diffarray = new Int8Array(randarray.length);
    for (index = 0; index < utf8bytes.length; index++) {
	diffarray[index] = randarray[index] - utf8bytes[index];
    }
    jsonstring = JSON.stringify({rand: Array.from(randarray), diff: Array.from(diffarray)})
    passwordbytes = utf8encoder.encode(jsonstring);
    return passwordbytes.toBase64();
}

let genAuth = function() {
    adminusername = document.getElementById("AdminUsername")
    adminpassword = document.getElementById("AdminPassword")
    if (adminusername === undefined || adminusername === null || adminusername.value.length <= 0 ||
	adminpassword === undefined || adminpassword === null || adminpassword.value.length <= 0) {
	console.log("empty adminusername or adminpassword");
	return "";
    }
    authstring = adminusername.value + ":" + encodePassword(adminpassword.value);
    textencoder = new TextEncoder();
    bytes = textencoder.encode(authstring);
    return bytes.toBase64();
}
