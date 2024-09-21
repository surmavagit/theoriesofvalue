const sortOptions = document.getElementById("sort-select");
if (sortOptions) {
    sortOptions.addEventListener("change", sortBy);

    function sortBy() {
        const field = sortOptions.value;
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
