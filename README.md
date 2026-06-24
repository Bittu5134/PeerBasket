<div align="center">
  <a href="https://peerbasket.bittu.dev/">
    <img src="./public/banner.webp" alt="PeerBasket Banner" width="100%" style="border-radius: 12px;" />
  </a>

  <br />
  <br />

  **A hassle-free, zero-config, lobby-based peer discovery API for PeerJS.**

  [![GitHub Release](https://img.shields.io/github/v/release/Bittu5134/PeerBasket?style=flat-square&color=emerald)](https://github.com/Bittu5134/PeerBasket/releases)
  [![Go Report Card](https://goreportcard.com/badge/github.com/Bittu5134/PeerBasket?style=flat-square)](https://goreportcard.com/report/github.com/Bittu5134/PeerBasket)
  [![License](https://img.shields.io/github/license/Bittu5134/PeerBasket?style=flat-square&color=blue)](LICENSE)
  [![Discord](https://img.shields.io/discord/877508637814300683?label=Discord&logo=discord&style=flat-square)](https://discord.gg/CZdNvKaNNr)
  [![Patreon](https://img.shields.io/badge/Patreon-Support-coral?style=flat-square&logo=patreon)](https://www.patreon.com/cw/LazyBittu)
</div>

---

## 🧺 What is PeerBasket?

[PeerJS](https://peerjs.com/) lets two browsers open a direct WebRTC peer-to-peer connection, but it has no built-in way for two peers to find each other in the first place (unless they share IDs out-of-band).

**PeerBasket solves this.** You send your Peer ID to a "basket" (a shared text string like `room-101`), and PeerBasket sends back the IDs of everyone else currently in that basket. No complex signaling infrastructure required.

A free, hosted public instance is running right now at: **[`peerbasket.bittu.dev`](https://peerbasket.bittu.dev/)**

## ⚡ Quickstart

Using PeerBasket is as simple as a single `fetch` request. Send your ID, get the other IDs back, and connect.

```javascript
const peer = new Peer(); // from the peerjs package

peer.on('open', async (myId) => {
  // 1. Register your ID in a basket
  const res = await fetch('https://peerbasket.bittu.dev/basket/my-awesome-room', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ peer_id: myId })
  });

  // 2. Get everyone else's IDs
  const { peers } = await res.json();

  // 3. Connect to them!
  peers
    .filter(id => id !== myId)
    .forEach(id => peer.connect(id));
});
```

> [!NOTE]
> The public API is rate-limited to 20 requests per minute. Polling **every 10 seconds** is the recommended sweet spot.

## 📚 Examples & Documentation

We don't want to clutter this README! We've built out a complete suite of copy-pasteable examples showing you how to build real apps with PeerBasket.

Check out the **[`./examples`](./examples)** folder for:

* 💬 **Group Chat Mesh:** Full multi-peer broadcast messaging.
* 📁 **File Transfer:** Sending raw binary files/Blobs directly between browsers.
* 🎥 **Video Calling:** Exchanging media streams.
* ⚛️ **React Matchmaking:** Using PeerBasket in a modern React/JSX frontend.
* 💻 **Terminal:** Using PeerBasket with Node.js.

For full API schema documentation, security considerations, and best practices, visit the **[official documentation website](https://peerbasket.bittu.dev/)**.

## 🛠️ Self-Hosting

Want to run your own private instance of PeerBasket instead of using the public API? It's incredibly lightweight.

### Prerequisite

PeerBasket uses **Redis** to instantly manage expiring heartbeats and cluster state. You must have Redis running locally (default: `localhost:6379`) or accessible via the `REDIS_ADDR` environment variable.

### Method 1: The Easy Way (Pre-built Binaries) ✨

You don't even need to install Go! We compile single-file, zero-dependency executables for Windows, macOS, and Linux.

1. Go to the **[Releases Tab](https://github.com/Bittu5134/PeerBasket/releases)**.
2. Download the binary for your operating system.
3. Run the executable. It will automatically host the API and serve the dashboard on port `8080`.

### Method 2: Build from Source

If you prefer to compile it yourself, you will need [Go](https://golang.org/dl/) installed.

```bash
# Clone the repo
git clone https://github.com/Bittu5134/PeerBasket.git
cd PeerBasket

# Install dependencies
go mod tidy

# Run the server
go run .
```

*Optional environment variables: `PORT=3000`, `REDIS_ADDR=127.0.0.1:6379`*

## 🤝 Support & Links

PeerBasket is built solo by **[Bittu](https://bittu.dev/)** and provided completely open-source with no strings attached.

If this tool saved you the headache of building a signaling server from scratch, consider supporting the project:

* ⭐ **Star this repository**
* 💖 **[Support me on Patreon](https://www.patreon.com/cw/LazyBittu)**
* 💬 **[Join the Discord Server](https://discord.gg/CZdNvKaNNr)**

---