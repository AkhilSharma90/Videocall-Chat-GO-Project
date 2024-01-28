// Get references to HTML elements
var msg = document.getElementById("msg");
var log = document.getElementById("log");

// Initialize slideOpen variable
var slideOpen = false;

// Toggle the visibility of the chat content
function slideToggle() {
    var chat = document.getElementById('chat-content');
    if (slideOpen) {
        chat.style.display = 'none';
        slideOpen = false;
    } else {
        chat.style.display = 'block'
        document.getElementById('chat-alert').style.display = 'none';
        document.getElementById('msg').focus();
        slideOpen = true
    }
}

// Append a new log item and handle scrolling
function appendLog(item) {
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
        log.scrollTop = log.scrollHeight - log.clientHeight;
    }
}

// Get the current time in HH:mm format
function currentTime() {
    var date = new Date;
    var hour = date.getHours();
    var minute = date.getMinutes();
    if (hour < 10) {
        hour = "0" + hour
    }
    if (minute < 10) {
        minute = "0" + minute
    }
    return hour + ":" + minute
}

// Handle form submission
document.getElementById("form").onsubmit = function () {
    // Check if WebSocket is available and message is not empty
    if (!chatWs) {
        return false;
    }
    if (!msg.value) {
        return false;
    }
    // Send message through WebSocket
    chatWs.send(msg.value);
    // Clear the message input
    msg.value = "";
    return false;
};

// Connect to the chat WebSocket
function connectChat() {
    chatWs = new WebSocket(ChatWebsocketAddr)

    // Handle WebSocket close event
    chatWs.onclose = function (evt) {
        console.log("websocket has closed")
        document.getElementById('chat-button').disabled = true
        // Reconnect after a delay
        setTimeout(function () {
            connectChat();
        }, 1000);
    }

    // Handle WebSocket message event
    chatWs.onmessage = function (evt) {
        var messages = evt.data.split('\n');
        // Display chat alert if the chat is not open
        if (slideOpen == false) {
            document.getElementById('chat-alert').style.display = 'block'
        }
        // Iterate through messages and display them in the log
        for (var i = 0; i < messages.length; i++) {
            var item = document.createElement("div");
            item.innerText = currentTime() + " - " + messages[i];
            appendLog(item);
        }
    }

    // Handle WebSocket error event
    chatWs.onerror = function (evt) {
        console.log("error: " + evt.data)
    }

    // Enable chat button after a delay if WebSocket is open
    setTimeout(function () {
        if (chatWs.readyState === WebSocket.OPEN) {
            document.getElementById('chat-button').disabled = false
        }
    }, 1000);
}

// Call the connectChat function to initiate the WebSocket connection
connectChat();
