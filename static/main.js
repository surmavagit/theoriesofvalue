const sortOptions = document.getElementById("sort-select");
sortOptions.addEventListener("change", sortBy);

function sortBy() {
    const field = sortOptions.value;

    const list = document.getElementById('author-list');

    [...list.children]
        .sort((a, b) => a.dataset[field] > b.dataset[field] ? 1 : -1)
        .forEach(node => {
            list.appendChild(node);
        });
}
