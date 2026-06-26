<div align="center">

<a href="https://peerbasket.bittu.dev">
    <img src="./public/banner.webp" alt="PeerBasket logo" title="PeerBasket logo" width="800"/>
</a>
<br />

# PeerBasket

### A hassle-free, lobby-based PeerJS discovery server.

[![API Status](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fpeerbasket.bittu.dev%2Fping&query=%24.status&style=flat-square&label=API+Status&color=success)](https://peerbasket.bittu.dev/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Bittu5134/PeerBasket?style=flat-square&color=00ADD8&logo=go&logoColor=white)](https://github.com/Bittu5134/PeerBasket)
[![GitHub Release](https://img.shields.io/github/v/release/Bittu5134/PeerBasket?style=flat-square&color=purple)](https://github.com/Bittu5134/PeerBasket/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/bittu5134/peerbasket/deploy.yml?style=flat-square&label=build)](https://github.com/Bittu5134/PeerBasket/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bittu5134/PeerBasket?style=flat-square)](https://goreportcard.com/report/github.com/Bittu5134/PeerBasket)
[![License: MIT](https://img.shields.io/github/license/Bittu5134/PeerBasket?style=flat-square&color=blue)](/LICENSE)
[![Discord](https://img.shields.io/discord/877508637814300683?label=discord&logo=discord&style=flat-square&color=7289da)](https://discord.gg/CZdNvKaNNr)

</div>

---

### The Problem
WebRTC enables two browsers to connect directly, but only after they exchange Peer IDs. WebRTC has no built-in way for peers to find each other in the first place, forcing developers to configure stateful signaling servers or databases just to bootstrap a connection.

### The Solution
PeerBasket solves this with a stateless, zero-configuration HTTP API. You POST your Peer ID to a room name (a basket), and receive a list of other active Peer IDs currently in it. These IDs are then passed to PeerJS to establish direct WebRTC connections.

### Key Features
* **Zero Config**: No signups, databases, or room setups. Just call the API endpoint.
* **Auto-Pruning**: Inactive peer heartbeats are automatically purged after 30 seconds to clean up empty lobbies.
* **Hosted Public Node**: A free public instance runs at `https://peerbasket.bittu.dev`.

For full rate limits, security details, and API schemas, refer to the [API Documentation](https://peerbasket.bittu.dev).

---

### Quickstart

Register your Peer ID and fetch other active peers in a basket with one POST request:

```javascript
const response = await fetch('https://peerbasket.bittu.dev/basket/my-room-id', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ peer_id: 'my-peerjs-id' })
});
const { peers } = await response.json();

// Connect to everyone else in the room
peers
  .filter(id => id !== 'my-peerjs-id')
  .forEach(id => peer.connect(id));
```

To see client integrations, check out the [Examples Directory](./examples):

---

### Self-Hosting

Requires **Redis** (running locally or accessible via `REDIS_ADDR` environment variable).

#### Option A: Using Release Binaries (Recommended)
1. Download the pre-compiled binary for your system from the [Releases Page](https://github.com/Bittu5134/PeerBasket/releases).
2. Create a `.env` file in the same directory:
   ```env
   PORT=8080
   GIN_MODE=release
   REDIS_ADDR=127.0.0.1:6379
   ```
3. Run the binary:
   * **Linux / macOS**:
     ```bash
     chmod +x peerbasket-linux-amd64 && ./peerbasket-linux-amd64
     ```
   * **Windows**:
     ```powershell
     .\peerbasket-windows-amd64.exe
     ```

#### Option B: Building from Source
1. **Prerequisite**: Install Go 1.26+.
2. Clone, install dependencies, and run:
   ```bash
   git clone https://github.com/Bittu5134/PeerBasket.git
   cd PeerBasket
   go mod tidy
   go run .
   ```

---

### Support & Sponsorship
If PeerBasket helps your project, please consider supporting development costs:
* **Patreon**: Support directly via [Patreon](https://www.patreon.com/cw/LazyBittu).
* **Community**: Join discussions on [Discord](https://discord.gg/CZdNvKaNNr).

---

### License
Distributed under the MIT License. See [LICENSE](./LICENSE) for details.
