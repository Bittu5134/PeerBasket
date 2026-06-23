// errors and reconnection
//
// This file is not a demo app. It's a reference for the failure modes
// you'll actually hit when using PeerBasket and PeerJS together, and a
// reusable helper, robustJoinBasket(), that handles all of them. Every
// other example in this folder uses a simplified version of this logic
// inline; this file is where the full version lives, with comments
// explaining WHY each branch exists.
//
// Run it directly to see it in action:
//   node main.js

import { Peer } from "peerjs";

const BASKET_ID = "peerbasket-demo-reconnect-reference";

// ---------------------------------------------------------------------
// Failure mode 1: PeerBasket is unreachable or returns an error.
//
// fetch() throws for network-level failures (offline, DNS, CORS) and
// resolves normally (without throwing) for HTTP error statuses, so
// both cases need to be checked separately. A 429 specifically means
// you're polling faster than PeerBasket's 20 requests/minute limit.
// ---------------------------------------------------------------------

async function joinBasketOnce(peer, basketId) {
  let response;
  try {
    response = await fetch(`https://peerbasket.bittu.dev/basket/${basketId}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ peer_id: peer.id }),
    });
  } catch (networkErr) {
    return { ok: false, reason: "network", error: networkErr };
  }

  if (response.status === 429) {
    return { ok: false, reason: "rate-limited" };
  }
  if (!response.ok) {
    return { ok: false, reason: "http-error", status: response.status };
  }

  const body = await response.json();
  return { ok: true, peers: body.peers };
}

// ---------------------------------------------------------------------
// Failure mode 2: a peer ID PeerBasket gave us is stale.
//
// PeerBasket only knows a peer was reachable as of its last heartbeat,
// up to 30 seconds ago. The peer behind that ID may have already
// closed its tab, lost its network connection, or crashed. The only
// reliable way to know if a peer is actually still there is to try to
// connect and see whether the connection opens.
// ---------------------------------------------------------------------

function connectWithTimeout(peer, remoteId, timeoutMs = 8000) {
  return new Promise((resolve) => {
    const conn = peer.connect(remoteId);
    let settled = false;

    const timer = setTimeout(() => {
      if (settled) return;
      settled = true;
      conn.close();
      resolve({ ok: false, reason: "timeout", peerId: remoteId });
    }, timeoutMs);

    conn.on("open", () => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      resolve({ ok: true, conn });
    });

    conn.on("error", (err) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      resolve({ ok: false, reason: "connection-error", peerId: remoteId, error: err });
    });
  });
}

// ---------------------------------------------------------------------
// Failure mode 3: the PeerJS broker connection itself drops.
//
// new Peer() talks to PeerJS's own cloud broker to do the initial
// WebRTC handshake. That connection can drop independently of
// PeerBasket and independently of any individual peer-to-peer
// connection. PeerJS emits a "disconnected" event when this happens,
// and exposes peer.reconnect() to try to restore the same peer ID
// without destroying everything and starting over.
//
// peer.on("error", ...) catches a broader set of problems, including
// some that ARE fatal (like "browser-incompatible"); reconnect() is
// only appropriate for the "disconnected" case.
// ---------------------------------------------------------------------

function createResilientPeer() {
  const peer = new Peer();

  peer.on("disconnected", () => {
    console.warn("Lost connection to the PeerJS broker, attempting to reconnect...");
    // A short delay avoids hammering the broker in a tight loop if it's
    // having a bad moment.
    setTimeout(() => {
      if (!peer.destroyed) peer.reconnect();
    }, 1000);
  });

  peer.on("error", (err) => {
    // "browser-incompatible", "ssl-unavailable", and similar are not
    // recoverable by reconnecting; log them distinctly so they don't
    // get mistaken for the transient "disconnected" case above.
    console.error(`PeerJS error (${err.type}):`, err.message);
  });

  return peer;
}

// ---------------------------------------------------------------------
// Putting it together: a poll loop that survives all three failure
// modes above without crashing or getting stuck.
// ---------------------------------------------------------------------

async function robustJoinBasket(peer, basketId, onConnection) {
  const result = await joinBasketOnce(peer, basketId);

  if (!result.ok) {
    // Every failure reason here is meant to be silently retried on the
    // next poll. We only log it, since a single failed poll is normal
    // and expected (a flaky network, a momentary rate limit, etc).
    console.warn(`Basket poll failed (${result.reason}), will retry`);
    return;
  }

  const candidates = result.peers.filter((id) => id !== peer.id);

  for (const id of candidates) {
    const outcome = await connectWithTimeout(peer, id);
    if (outcome.ok) {
      onConnection(outcome.conn);
    } else {
      // Don't let one stale peer ID block the rest of the basket from
      // being processed.
      console.warn(`Could not connect to ${id} (${outcome.reason}), skipping`);
    }
  }
}

// ---------------------------------------------------------------------
// Demo: run the loop against a real basket so you can see it working.
// ---------------------------------------------------------------------

const peer = createResilientPeer();
const connections = new Map();

peer.on("open", (myId) => {
  console.log(`My peer ID is ${myId}`);

  const poll = () => robustJoinBasket(peer, BASKET_ID, (conn) => {
    if (connections.has(conn.peer)) return;
    connections.set(conn.peer, conn);
    console.log(`Connected to ${conn.peer}`);
    conn.on("close", () => connections.delete(conn.peer));
  });

  poll();
  setInterval(poll, 10_000);
});

peer.on("connection", (conn) => {
  connections.set(conn.peer, conn);
  console.log(`${conn.peer} connected to us`);
  conn.on("close", () => connections.delete(conn.peer));
});