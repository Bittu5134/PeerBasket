// basic-connect.js
//
// A command-line peer using PeerBasket and PeerJS's native Node.js
// support. Run this script in two separate terminals; each instance
// will find the other through PeerBasket and let you type messages
// back and forth.
//
// NOTE: PeerJS's Node.js/Bun/Deno support is currently in beta. If you
// hit WebRTC-related errors on your platform, the browser examples would
// be a better fit for you.
//
// Setup:
//   npm install peerjs
//
// Run (in two separate terminals):
//   node basic-connect.js

import { Peer } from "peerjs";
import readline from "node:readline";

// Both processes must use the same basket name to find each other.
const BASKET_ID = "peerbasket-demo-cli-basic-connect";
const POLL_INTERVAL_MS = 10_000;

// peerId -> open DataConnection.
const connections = new Map();

const peer = new Peer();

peer.on("open", (myId) => {
  console.log(`My peer ID is ${myId}`);
  console.log("Type a message and press Enter to broadcast it.\n");

  joinBasket();
  setInterval(joinBasket, POLL_INTERVAL_MS);
});

// Fired when another peer connects to us first.
peer.on("connection", (conn) => setupConnection(conn));

peer.on("error", (err) => {
  console.error(`PeerJS error: ${err.type}`);
});

async function joinBasket() {
  let response;
  try {
    response = await fetch(`https://peerbasket.bittu.dev/basket/${BASKET_ID}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ peer_id: peer.id }),
    });
  } catch (networkErr) {
    console.error("Could not reach PeerBasket, will retry next interval");
    return;
  }

  if (!response.ok) {
    console.error(`PeerBasket responded with HTTP ${response.status}`);
    return;
  }

  const { peers } = await response.json();

  peers
    .filter((id) => id !== peer.id)
    .filter((id) => !connections.has(id))
    .forEach((id) => setupConnection(peer.connect(id)));
}

function setupConnection(conn) {
  connections.set(conn.peer, conn);

  conn.on("open", () => {
    console.log(`Connected to ${conn.peer}`);
  });

  conn.on("data", (data) => {
    console.log(`${conn.peer}: ${data}`);
  });

  conn.on("close", () => {
    connections.delete(conn.peer);
    console.log(`${conn.peer} disconnected`);
  });

  conn.on("error", () => {
    connections.delete(conn.peer);
  });
}

// Read lines from stdin and broadcast each one to every open connection.
const rl = readline.createInterface({ input: process.stdin });
rl.on("line", (text) => {
  if (!text.trim()) return;
  connections.forEach((conn) => {
    if (conn.open) conn.send(text);
  });
});