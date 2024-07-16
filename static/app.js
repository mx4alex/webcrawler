document.getElementById('search-form').addEventListener('submit', function(event) {
    event.preventDefault();

    const searchInput = document.getElementById('search-input').value;
    const resultsDiv = document.getElementById('results');
    resultsDiv.innerHTML = '';

    console.log(`Searching for: ${searchInput}`);

    fetch(`/webcrawler/search/${searchInput}`)
        .then(response => response.json())
        .then(data => {
            console.log('Received data:', data);

            if (data.urls.length === 0) {
                resultsDiv.innerHTML = '<p>No results found.</p>';
            } else {
                data.urls.forEach(url => {
                    const div = document.createElement('div');
                    div.classList.add('result-item');
                    div.innerHTML = `<a href="${url}" target="_blank">${url}</a>`;
                    resultsDiv.appendChild(div);
                });
            }
        })
        .catch(error => {
            console.error('Error fetching search results:', error);
            resultsDiv.innerHTML = '<p>Error fetching search results.</p>';
        });
});
