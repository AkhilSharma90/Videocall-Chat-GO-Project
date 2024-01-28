// Function to copy text to clipboard
function copyToClipboard(text) {
	if (window.clipboardData && window.clipboardData.setData) {
		clipboardData.setData("Text", text);
		// Show notification using SweetAlert
		return Swal.fire({
			position: 'top-end',
			text: "Copied",
			showConfirmButton: false,
			timer: 1000,
			width: '150px'
		})
	} else if (document.queryCommandSupported && document.queryCommandSupported("copy")) {
		// Create a temporary textarea element to copy text
		var textarea = document.createElement("textarea");
		textarea.textContent = text;
		textarea.style.position = "fixed";
		document.body.appendChild(textarea);
		textarea.select();
		try {
			// Execute copy command
			document.execCommand("copy");
			// Show notification using SweetAlert
			return Swal.fire({
				position: 'top-end',
				text: "Copied",
				showConfirmButton: false,
				timer: 1000,
				width: '150px'
			})
		} catch (ex) {
			console.warn("Copy to clipboard failed.", ex);
			return false;
		} finally {
			document.body.removeChild(textarea);
		}
	}
}

// Event listener when DOM content is loaded
document.addEventListener('DOMContentLoaded', () => {
	// Add click event to close notification
	(document.querySelectorAll('.notification .delete') || []).forEach(($delete) => {
		const $notification = $delete.parentNode;

		$delete.addEventListener('click', () => {
			$notification.style.display = 'none'
		});
	});
});

// Function to connect to the chat
function connect(stream) {
	// Display elements
	document.getElementById('peers').style.display = 'block'
	document.getElementById('chat').style.display = 'flex'
	document.getElementById('noperm').style.display = 'none'

	// Create an RTCPeerConnection
	let pc = new RTCPeerConnection({
		iceServers: [{
				'urls': 'stun:turn.videochat:3478',
			},
			{
				'urls': 'turn:turn.videochat:3478',
				'username': 'akhil',
				'credential': 'akhil',
			}
		]
	})

	// Event handler for remote tracks
	pc.ontrack = function (event) {
		if (event.track.kind === 'audio') {
			return
		}

		// Create elements for displaying remote video
		col = document.createElement("div")
		col.className = "column is-6 peer"
		let el = document.createElement(event.track.kind)
		el.srcObject = event.streams[0]
		el.setAttribute("controls", "true")
		el.setAttribute("autoplay", "true")
		el.setAttribute("playsinline", "true")
		col.appendChild(el)
		document.getElementById('noone').style.display = 'none'
		document.getElementById('nocon').style.display = 'none'
		document.getElementById('videos').appendChild(col)

		// Event handlers for track mute and removal
		event.track.onmute = function (event) {
			el.play()
		}

		event.streams[0].onremovetrack = ({
			track
		}) => {
			if (el.parentNode) {
				el.parentNode.remove()
			}
			if (document.getElementById('videos').childElementCount <= 3) {
				document.getElementById('noone').style.display = 'grid'
				document.getElementById('noonein').style.display = 'grid'
			}
		}
	}

	// Add tracks from the stream to the connection
	stream.getTracks().forEach(track => pc.addTrack(track, stream))

	// Create a WebSocket for signaling
	let ws = new WebSocket(RoomWebsocketAddr)
	pc.onicecandidate = e => {
		if (!e.candidate) {
			return
		}

		// Send ICE candidate through WebSocket
		ws.send(JSON.stringify({
			event: 'candidate',
			data: JSON.stringify(e.candidate)
		}))
	}

	// Event handler for WebSocket error
	ws.addEventListener('error', function (event) {
		console.log('error: ', event)
	})

	// Event handler for WebSocket close
	ws.onclose = function (evt) {
		console.log("websocket has closed")
		pc.close();
		pc = null;
		pr = document.getElementById('videos')
		while (pr.childElementCount > 3) {
			pr.lastChild.remove()
		}
		document.getElementById('noone').style.display = 'none'
		document.getElementById('nocon').style.display = 'flex'
		// Reconnect after a delay
		setTimeout(function () {
			connect(stream);
		}, 1000);
	}

	// Event handler for WebSocket messages
	ws.onmessage = function (evt) {
		let msg = JSON.parse(evt.data)
		if (!msg) {
			return console.log('failed to parse msg')
		}

		// Handle different types of messages
		switch (msg.event) {
			case 'offer':
				let offer = JSON.parse(msg.data)
				if (!offer) {
					return console.log('failed to parse answer')
				}
				// Set remote description and create answer
				pc.setRemoteDescription(offer)
				pc.createAnswer().then(answer => {
					pc.setLocalDescription(answer)
					// Send answer through WebSocket
					ws.send(JSON.stringify({
						event: 'answer',
						data: JSON.stringify(answer)
					}))
				})
				return

			case 'candidate':
				let candidate = JSON.parse(msg.data)
				if (!candidate) {
					return console.log('failed to parse candidate')
				}

				// Add ICE candidate to the connection
				pc.addIceCandidate(candidate)
		}
	}

	// Event handler for WebSocket error
	ws.onerror = function (evt) {
		console.log("error: " + evt.data)
	}
}

// Get user media with specified constraints
navigator.mediaDevices.getUserMedia({
		video: {
			width: {
				max: 1280
			},
			height: {
				max: 720
			},
			aspectRatio: 4 / 3,
			frameRate: 30,
		},
		audio: {
			sampleSize: 16,
			channelCount: 2,
			echoCancellation: true
		}
	})
	.then(stream => {
		// Set local video element source
		document.getElementById('localVideo').srcObject = stream
		// Connect to the chat
		connect(stream)
	}).catch(err => console.log(err))
