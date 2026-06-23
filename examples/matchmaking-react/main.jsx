// Matchmaking in React
//
// A React component that pairs the current browser with exactly one
// other browser using PeerBasket, then opens a PeerJS data connection
// between them. Once paired, it stops polling PeerBasket entirely,
// which keeps long-running matched sessions well under the rate limit.
//
// This is "1:1 random matchmaking": whoever is in the basket together
// gets paired, in whatever order PeerBasket happens to return them.
// There's no queueing or "next" button here; if you want users to be
// able to skip a match and find a new one, see the note at the bottom
// of this file.
//
// Requires the "peerjs" package:
//   npm install peerjs
//
// Usage:
//   <Matchmaking basketId="my-app-matchmaking-v1" />

import { useEffect, useRef, useState } from "react";
import Peer from "peerjs";

const POLL_INTERVAL_MS = 10_000;

export default function Matchmaking({ basketId }) {
  const [myId, setMyId] = useState(null);
  const [status, setStatus] = useState("connecting");
  const [messages, setMessages] = useState([]);
  const [draft, setDraft] = useState("");

  // Refs hold values that the polling/event callbacks need to read but
  // that should NOT trigger a re-render or a stale closure when they
  // change. Using React state for these would either cause unnecessary
  // re-renders or capture an outdated value inside setInterval.
  const peerRef = useRef(null);
  const connectionRef = useRef(null);
  const pollTimerRef = useRef(null);

  useEffect(() => {
    const peer = new Peer();
    peerRef.current = peer;

    peer.on("open", (id) => {
      setMyId(id);
      setStatus("waiting for a match");
      pollOnce(); // try immediately, then on an interval
      pollTimerRef.current = setInterval(pollOnce, POLL_INTERVAL_MS);
    });

    // Handles the case where the OTHER peer is the one who finds the
    // match first and connects to us.
    peer.on("connection", (conn) => {
      if (connectionRef.current) {
        conn.close(); // already matched, reject anyone else
        return;
      }
      attachConnection(conn);
    });

    peer.on("error", (err) => {
      setStatus(`error: ${err.type}`);
    });

    async function pollOnce() {
      if (connectionRef.current) return; // already matched, stop polling logic

      let response;
      try {
        response = await fetch(`https://peerbasket.bittu.dev/basket/${basketId}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ peer_id: peer.id }),
        });
      } catch {
        return; // network hiccup, just try again next interval
      }
      if (!response.ok) return;

      const { peers } = await response.json();
      const candidate = peers.find((id) => id !== peer.id);

      if (candidate && !connectionRef.current) {
        attachConnection(peer.connect(candidate));
      }
    }

    function attachConnection(conn) {
      connectionRef.current = conn;
      clearInterval(pollTimerRef.current); // matched, no need to keep polling

      conn.on("open", () => setStatus(`matched with ${conn.peer}`));

      conn.on("data", (text) => {
        setMessages((prev) => [...prev, { from: conn.peer, text }]);
      });

      conn.on("close", () => {
        connectionRef.current = null;
        setStatus("match ended");
      });
    }

    // Cleanup: runs if the component unmounts, or if basketId changes
    // and this effect re-runs. Without this, leaving the page would
    // leave a stale connection and a running poll timer behind.
    return () => {
      clearInterval(pollTimerRef.current);
      connectionRef.current?.close();
      peer.destroy();
    };
  }, [basketId]);

  function sendMessage() {
    const conn = connectionRef.current;
    if (!conn?.open || !draft.trim()) return;

    conn.send(draft);
    setMessages((prev) => [...prev, { from: "me", text: draft }]);
    setDraft("");
  }

  return (
    <div>
      <p>My peer ID: {myId ?? "connecting..."}</p>
      <p>Status: {status}</p>

      <ul>
        {messages.map((m, i) => (
          <li key={i}>
            {m.from}: {m.text}
          </li>
        ))}
      </ul>

      <input
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && sendMessage()}
        placeholder="Type a message and press Enter"
      />
    </div>
  );
}

// ---------------------------------------------------------------------
// Note on "skip this match, find another":
//
// To let a user leave their current match and look for a new one,
// close connectionRef.current, set it back to null, and restart the
// polling interval. The basket itself doesn't need to change; the next
// poll will simply look for whoever else happens to be registered in
// it at that moment, possibly including the same peer again if no one
// else has joined yet.
// ---------------------------------------------------------------------