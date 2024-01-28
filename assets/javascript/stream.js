function connectStream() {
	// Show video container and chat when connecting to the stream
	document.getElementById('peers').style.display = 'block';
	document.getElementById('chat').style.display = 'flex';

	// Create a new RTCPeerConnection with ICE servers configuration
	let pc = new RTCPeerConnection({
		iceServers: [
			{
				'urls': 'stun:turn.videochat:3478',
			},
			{
				'urls': 'turn:turn.videochat:3478',
				'username': 'akhil',
				'credential': 'akhil',
			}
		]
	});

	// Event handler when a media track is added to the connection
	pc.ontrack = function (event) {
		// Skip audio tracks
		if (event.track.kind === 'audio') {
			return;
		}

		// Create HTML elements for video display
		let col = document.createElement("div");
		col.className = "column is-6 peer";
		let el = document.createElement(event.track.kind);
		el.srcObject = event.streams[0];
		el.setAttribute("controls", "true");
		el.setAttribute("autoplay", "true");
		el.setAttribute("playsinline", "true");

		// Attempt to play the video, 
