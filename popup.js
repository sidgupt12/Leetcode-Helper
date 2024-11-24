document.getElementById('getHint').addEventListener('click', function () {
    const urlElement = document.getElementById('url');
    const loaderContainer = document.querySelector('.loader-container');
    const userDescription = document.getElementById('userDescription').value;
    
    // Show loader, hide previous results
    loaderContainer.classList.add('active');
    urlElement.classList.remove('active');
    urlElement.innerHTML = '';

    chrome.tabs.query({ active: true, currentWindow: true }, function (tabs) {
        const url = tabs[0].url;
        console.log('URL from current tab:', url);

        fetch('http://127.0.0.1:8080/capture-url', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                 url: url,
                 description: userDescription
                 })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log("Server response:", data);
            let responseHTML = "";

            if (data.Problem_Title) {
                responseHTML += `
                    <div class="response-item">
                        <div class="title">Problem Title</div>
                        <div class="content">${data.Problem_Title}</div>
                    </div>`;
            }

            if (data.hint) {
                // Split the hint by numbered items and filter out empty strings
                const hints = data.hint.split(/\d+\.\s+/).filter(item => item.trim().length > 0);
                
                responseHTML += `
                    <div class="response-item">
                        <div class="title">Hints</div>
                        <div class="content">
                            <ul class="hint-list">`;
                
                hints.forEach(hint => {
                    responseHTML += `
                        <li class="hint-item">
                            <div class="hint-content">${hint.trim()}</div>
                        </li>`;
                });
                
                responseHTML += `
                            </ul>
                        </div>
                    </div>`;
            } else {
                responseHTML += `
                    <div class="response-item">
                        <div class="title">Hint</div>
                        <div class="content">No hint available.</div>
                    </div>`;
            }

            if (data.message) {
                responseHTML += `
                    <div class="response-item">
                        <div class="title">Message</div>
                        <div class="content">${data.message}</div>
                    </div>`;
            }

            // Hide loader and show results
            loaderContainer.classList.remove('active');
            urlElement.innerHTML = responseHTML;
            urlElement.classList.add('active');
        })
        .catch(err => {
            console.error("Fetch error:", err);
            // Hide loader and show error
            loaderContainer.classList.remove('active');
            urlElement.innerHTML = `
                <div class="response-item">
                    <div class="title">Error</div>
                    <div class="content">Failed to fetch hint. Please try again.</div>
                </div>`;
            urlElement.classList.add('active');
        });
    });
});