document.getElementById('getHint').addEventListener('click', function () {
    chrome.tabs.query({ active: true, currentWindow: true }, function (tabs) {
        const url = tabs[0].url;
        console.log('URL from current tab:', url);  // Log the URL

        fetch('http://127.0.0.1:8080/capture-url', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url: url })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log("Server response:", data);  // Log the server response

            // Get the single div element where all responses will be displayed
            const urlElement = document.getElementById('url');

            // Construct the message to display all responses
            let responseMessage = "";

            if (data.url) {
                responseMessage += `Problem Title: ${data.url}<br>`;
            }
            if (data.message) {
                responseMessage += `Message: ${data.message}<br>`;
            }
            if (data.hint) {
                responseMessage += `Hint: ${data.hint}<br>`;
            } else {
                responseMessage += "No hint available.<br>";
            }

            // Update the content of the single div with the formatted response message
            urlElement.innerHTML = responseMessage;
        })
        .catch(err => console.error("Fetch error:", err));
    });
});