# PeerBasket Examples

These barebones, use-case driven examples demonstrate the direct value and specialty of **PeerBasket** (lobby-based discovery, matchmaking, and connection handshaking) separate from WebRTC infrastructure setup.

Every example connects to the hosted instance at `https://peerbasket.bittu.dev` by default, so they can be opened directly in your browser without setting up your own server.

---

## Example 1: 🤝 [1:1 Matchmaker Chat](./01-matchmaker-chat/index.html)
* **Usecase**: A private chat system.
* **PeerBasket Speciality**: Matches you up with a random partner. Once a connection is established, it **immediately stops polling** PeerBasket to respect server resources and prevent rate limits.

## Example 2: 📁 [P2P File Drop](./02-local-file-drop/index.html)
* **Usecase**: Local file sharing (AirDrop clone).
* **PeerBasket Speciality**: Serves as a dynamic peer registry. You see everyone currently in the room, select one, and send raw file buffers directly from browser to browser.

## Example 3: 🎥 [Group Video Room](./03-group-video-chat/index.html)
* **Usecase**: Multi-user video meetings (like Zoom or Google Meet).
* **PeerBasket Speciality**: Automatically builds a full mesh WebRTC video network. PeerBasket discovers all active users in the room, enabling each client to call all other peers dynamically, and automatically terminates/removes video feeds when peers stop sending heartbeats.

---

### Local Testing (Self-Hosted)
If you are running PeerBasket locally, you can test these examples against your local server:
1. Open the example's `index.html` file.
2. Change the `API_URL` variable at the top of the `<script>` tag:
   ```javascript
   const API_URL = "http://localhost:8080";
   ```
