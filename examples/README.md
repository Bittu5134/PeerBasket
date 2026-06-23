# PeerBasket Examples

Working examples across different environments and use cases. Every example connects to the hosted instance at `peerbasket.bittu.dev`, so you can run them immediately without setting up your own server.

## Quick start

Open any `index.html` file in a modern browser, or run a Node.js example with `npm install && node main.js`. Open the same example in two tabs or terminals to see two peers discover each other.

## Examples by use case

### Basic peer discovery

**Start here if you're new.** A peer registers with a basket and connects to one or every other peer in it.

- **[`basic-demo-browser`](./basic-demo-browser/)**: vanilla HTML/JS. Open `index.html` in two browser tabs. Type and send messages back and forth.
- **[`basic-demo-terminal`](./basic-demo-terminal/)**: Node.js CLI. Run `node main.js` in two terminals. Type and send messages back and forth.

### Group chat / mesh

Multiple peers in the same basket, all connected directly to each other. Every peer broadcasts to all peers.

- **[`basic-groupchat`](./basic-groupchat/)**: vanilla HTML/JS. Open `index.html` in three or more browser tabs to see it in action. Great for small multiplayer lobbies or collaborative tools.

### Random 1:1 matchmaking

A user joins a basket and connects to exactly one other user, then stops polling. Good for matchmaking or pairing for a task.

- **[`matchmaking-react`](./matchmaking-react/)**: React component. `<Matchmaking basketId="my-app-matches-v1" />`. Works with any React app.

### File transfer

Send a file directly from one browser to another. Uses binary data channels and handles the full send/receive loop.

- **[`file-transfer`](./file-transfer/)**: vanilla HTML/JS. Open `index.html` in two browser tabs, pick a file in one, and download it in the other.

### Video and voice calls

WebRTC media streams (audio/video) flowing directly between browsers. PeerBasket only finds the peer; PeerJS handles the media.

- **[`videocall`](./videocall/)**: vanilla HTML/JS. Open `index.html` in two browser tabs. Allows camera/microphone access in both, and a 1:1 video call connects automatically.
- **[`videocall-react`](./videocall-react/)**: React component. `<VideoCall basketId="my-app-calls-v1" />`. Handles cleanup (stops camera, closes streams) on unmount.

**Note:** Video call examples require `https://` or `localhost`. They will not work if served over plain `http` on a non-localhost address, or opened as a `file://` URL. This is a browser security restriction, not a PeerJS or PeerBasket limitation.

### Error handling and resilience

A reference for all the failure modes you'll actually hit in production: PeerBasket unreachable, rate limited, stale peer IDs, the PeerJS broker connection dropping.

- **[`error-cases`](./error-cases/)**: Node.js reference module with `robustJoinBasket()` helper and detailed comments on why each check exists. Run `node main.js` to see it connect to the demo basket. Read the code to understand what can go wrong.

## By environment

### Browser (vanilla HTML/JS)

No build step, no dependencies. PeerJS is loaded from a CDN with a single `<script>` tag. Open any `index.html` directly in a browser.

- `basic-demo-browser/`
- `basic-groupchat/`
- `file-transfer/`
- `videocall/`

### Node.js

Command-line peers using PeerJS's native Node.js support (currently in beta).

```bash
cd basic-demo-terminal
npm install
node main.js
```

```bash
cd error-cases
npm install
node main.js
```

**Note:** PeerJS's Node.js, Bun, and Deno support is beta. If you hit WebRTC errors on your platform, the browser examples are more battle-tested.

### React

Components ready to drop into a React app. Requires `npm install peerjs`.

```jsx
import Matchmaking from "./matchmaking-react/main.jsx";
import VideoCall from "./videocall-react/main.jsx";

export default function App() {
  return (
    <>
      <Matchmaking basketId="my-app-matches-v1" />
      <VideoCall basketId="my-app-calls-v1" />
    </>
  );
}
```

## Important notes

### Basket IDs are public

Every example uses a demo basket name like `peerbasket-demo-browser`. Anyone running these exact files lands in the same basket as you.

**For production:** replace all basket names with your own long, random, unguessable string. Treat a basket ID like an unlisted URL, not a password. See the main README's [Security Considerations](../README.md#security-considerations) section.

### Peers are discoverable, not verified

PeerBasket has no authentication. Anyone who knows your basket ID can join it and see who else is in it. If you need to restrict membership, verify identity at the application layer: a shared secret, an OAuth token, a handshake once the PeerJS connection opens, etc.

### Active ≠ reachable

A peer is "active" if it checked in with PeerBasket within the last 30 seconds. But it might have crashed 5 seconds ago, lost its network connection, or closed its tab. The only real liveness check is whether a PeerJS connection actually opens. Handle the case where it never does.

### Rate limits

PeerBasket allows 20 requests per minute. Polling faster than every 10 seconds will hit the limit. Every example respects this; if you modify one, keep the interval >= 10 seconds.