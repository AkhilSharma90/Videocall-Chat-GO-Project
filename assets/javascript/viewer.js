// Function to connect to the viewer WebSocket
function connectViewer() {
    // Get the viewer count element
    viewerCount = document.getElementById("viewer-count");
    
    // Create a new WebSocket for viewer count
    viewerWs = new WebSocket(ViewerWebsocketAddr);

    // Event handler for WebSocket close
    viewerWs.onclose = function (evt) {
        console.log("websocket has closed");
        // Reset viewer count to 0
        viewerCount.innerHTML = "0";
        // Reconnect after a delay
        setTimeout(function () {
            connectViewer();
        }, 1000);
    }

    // Event handler for WebSocket messages
    viewerWs.onmessage = function (evt) {
        // Parse the data received
        d = evt.data
        // Check if the data is an integer
        if (d === parseInt(d, 10)) {
            return
        }
        // Update viewer count in the HTML
        viewerCount.innerHTML = d;
    }

    // Event handler for WebSocket error
    viewerWs.onerror = function (evt) {
        console.log("error: " + evt.data)
    } 
}

// Call the connectViewer function to initiate the WebSocket connection
connectViewer();
