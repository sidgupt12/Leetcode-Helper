// Function to save API key
document.getElementById('saveApiKey').addEventListener('click', function() {
    const apiKey = document.getElementById('apiKeyInput').value.trim();
    console.log('Save API key clicked');
    
    if (apiKey) {
        chrome.storage.sync.set({ 'geminiApiKey': apiKey }, function() {
            const statusDiv = document.querySelector('.api-key-status');
            statusDiv.textContent = 'API key saved successfully!';
            statusDiv.style.color = '#4ade80';
            console.log('API key saved');
            setTimeout(() => {
                statusDiv.textContent = '';
            }, 3000);
        });
    } else {
        const statusDiv = document.querySelector('.api-key-status');
        statusDiv.textContent = 'Please enter an API key';
        statusDiv.style.color = '#ef4444';
        setTimeout(() => {
            statusDiv.textContent = '';
        }, 3000);
    }
});

// Load saved API key when popup opens
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM Content Loaded');
    chrome.storage.sync.get(['geminiApiKey'], function(result) {
        if (result.geminiApiKey) {
            document.getElementById('apiKeyInput').value = result.geminiApiKey;
            console.log('Loaded saved API key');
        }
    });
});

// Main function for getting hints
document.getElementById('getHint').addEventListener('click', function () {
    console.log('Get Hint clicked');
    const urlElement = document.getElementById('url');
    const loaderContainer = document.querySelector('.loader-container');
    const userDescription = document.getElementById('userDescription').value;
    
    // Show loader, hide previous results
    loaderContainer.classList.add('active');
    urlElement.classList.remove('active');
    urlElement.innerHTML = '';

    // Get the API key and make the request
    chrome.storage.sync.get(['geminiApiKey'], function(result) {       
        chrome.tabs.query({ active: true, currentWindow: true }, function (tabs) {
            const url = tabs[0].url;
            console.log('URL from current tab:', url);

            // Add error checking for URL
            if (!url.includes('leetcode.com/problems/')) {
                loaderContainer.classList.remove('active');
                urlElement.innerHTML = `
                    <div class="response-item">
                        <div class="title">Error</div>
                        <div class="content">Please navigate to a LeetCode problem page first.</div>
                    </div>`;
                urlElement.classList.add('active');
                return;
            }

            fetch('http://127.0.0.1:8080/capture-url', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    url: url,
                    description: userDescription,
                    apiKey: result.geminiApiKey
                })
            })
            .then(response => {
                console.log('Server response status:', response.status);
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
                            <div class="content">No hint available. Please ensure that correct API key is entered</div>
                        </div>`;
                }

                if (data.message) {
                    responseHTML += `
                        <div class="response-item">
                            <div class="title">Message</div>
                            <div class="content">${data.message}</div>
                        </div>`;
                }

                loaderContainer.classList.remove('active');
                urlElement.innerHTML = responseHTML;
                urlElement.classList.add('active');
            })
            .catch(err => {
                console.error("Fetch error:", err);
                loaderContainer.classList.remove('active');
                urlElement.innerHTML = `
                    <div class="response-item">
                        <div class="title">Error</div>
                        <div class="content">Failed to fetch please ensure you entered the correct API</div>
                    </div>`;
                urlElement.classList.add('active');
            });
        });
    });
});