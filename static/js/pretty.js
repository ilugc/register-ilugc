document.querySelectorAll("label").forEach((label) => {
    label.textContent = label.textContent.replace(/([a-z])([A-Z])/g, "$1 $2");
});
