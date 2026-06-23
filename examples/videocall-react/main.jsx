// VideoCall for React
//
// A React component for a 1:1 video/voice call, using PeerBasket to
// find the other peer and PeerJS's media-call API (peer.call) to
// stream audio/video directly between browsers.
//
// Requires the "peerjs" package:
//   npm install peerjs
//
// Requires https:// or localhost for camera/microphone access to work
// (a browser restriction, not a PeerJS or PeerBasket one).
//
// Usage:
//   <VideoCall basketId="my-app-video-room-42" />

import { useEffect, useRef, useState } from "react";
import Peer from "peerjs";

const POLL_INTERVAL_MS = 10_000;

export default function VideoCall({ basketId }) {
  const [status, setStatus] = useState("requesting camera access");

  const localVideoRef = useRef(null);
  const remoteVideoRef = useRef(null);
  const peerRef = useRef(null);
  const localStreamRef = useRef(null);
  const activeCallRef = useRef(null);
  const pollTimerRef = useRef(null);

  useEffect(() => {
    let cancelled = false; // guards against setting state after unmount

    async function start() {
      let stream;
      try {
        stream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: true,
        });
      } catch {
        if (!cancelled) setStatus("camera/microphone access was denied");
        return;
      }
      if (cancelled) {
        // Component unmounted while the permission prompt was open.
        stream.getTracks().forEach((t) => t.stop());
        return;
      }

      localStreamRef.current = stream;
      if (localVideoRef.current) localVideoRef.current.srcObject = stream;
      setStatus("waiting for a peer");

      const peer = new Peer();
      peerRef.current = peer;

      peer.on("open", (id) => {
        pollOnce(peer);
        pollTimerRef.current = setInterval(() => pollOnce(peer), POLL_INTERVAL_MS);
      });

      // Someone else found us in the basket and called us first.
      peer.on("call", (incomingCall) => {
        if (activeCallRef.current) {
          incomingCall.close(); // already on a call
          return;
        }
        incomingCall.answer(localStreamRef.current);
        attachCall(incomingCall);
      });

      peer.on("error", (err) => {
        if (!cancelled) setStatus(`error: ${err.type}`);
      });
    }

    async function pollOnce(peer) {
      if (activeCallRef.current) return;

      let response;
      try {
        response = await fetch(`https://peerbasket.bittu.dev/basket/${basketId}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ peer_id: peer.id }),
        });
      } catch {
        return;
      }
      if (!response.ok) return;

      const { peers } = await response.json();
      const other = peers.find((id) => id !== peer.id);

      if (other && !activeCallRef.current) {
        attachCall(peer.call(other, localStreamRef.current));
      }
    }

    function attachCall(call) {
      activeCallRef.current = call;
      if (!cancelled) setStatus(`connecting to ${call.peer}`);

      call.on("stream", (remoteStream) => {
        if (remoteVideoRef.current) remoteVideoRef.current.srcObject = remoteStream;
        if (!cancelled) setStatus(`in call with ${call.peer}`);
      });

      call.on("close", () => {
        activeCallRef.current = null;
        if (remoteVideoRef.current) remoteVideoRef.current.srcObject = null;
        if (!cancelled) setStatus("waiting for a peer");
      });
    }

    start();

    // Cleanup: stop the camera, end any active call, and destroy the
    // peer connection. Skipping this would leave the camera light on
    // and a stale connection running after the component unmounts.
    return () => {
      cancelled = true;
      clearInterval(pollTimerRef.current);
      activeCallRef.current?.close();
      localStreamRef.current?.getTracks().forEach((t) => t.stop());
      peerRef.current?.destroy();
    };
  }, [basketId]);

  return (
    <div>
      <p>Status: {status}</p>
      <video ref={localVideoRef} autoPlay muted playsInline width={320} />
      <video ref={remoteVideoRef} autoPlay playsInline width={320} />
    </div>
  );
}