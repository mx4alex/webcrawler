document.getElementById('search-form').addEventListener('submit', function(event) {
    event.preventDefault();

    const searchInput = document.getElementById('search-input').value;
    const resultsDiv = document.getElementById('results');
    resultsDiv.innerHTML = '';

    console.log(`Searching for: ${searchInput}`);

    fetch(`/webcrawler/search/${encodeURIComponent(searchInput)}`)
        .then(response => response.json())
        .then(data => {
            console.log('Received data:', data);

            if (data.length === 0) {
                resultsDiv.innerHTML = '<p>No results found.</p>';
            } else {
                data.forEach(item => {
                    const div = document.createElement('div');
                    div.classList.add('result-item');
                    div.innerHTML = `<a href="${item.url}" target="_blank">${item.url}</a><p>${item.text}</p>`;
                    resultsDiv.appendChild(div);
                });
            }
            window.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        })
        .catch(error => {
            console.error('Error fetching search results:', error);
            resultsDiv.innerHTML = '<p>Error fetching search results.</p>';
        });
});
