const radios = document.getElementsByName("sort-radio");

if (radios.length > 1) {
    radios.forEach(radio => radio.addEventListener("change",
        function() {
            sortBy(radio.getAttribute("value"));
        }
    ));

    function sortBy(field) {
        const list = document.getElementById('author-list');

        function compareFields(a, b) {
            let aVal = a.dataset[field], bVal = b.dataset[field];
            if (!isNaN(aVal)) {
                aVal = +aVal;
                bVal = +bVal;
            }
            return aVal > bVal ? 1 : -1;
        }

        [...list.children]
            .sort(compareFields)
            .forEach(node => {
                list.appendChild(node);
            });
    }
}
