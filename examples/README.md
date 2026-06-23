# PeerBasket examples

Working examples for every common PeerBasket use case, across four environments. Every example uses real calls to the hosted instance at `peerbasket.bittu.dev`, so you can run them as they are, without setting up your own server.

All examples use a basket name like `peerbasket-demo-basic-connect`. That name is public, so anyone running these exact files will land in the same basket as you. For real projects, replace it with your own long, random, unguessable string. See the [Security Considerations](../README.md#security-considerations) section of the main README for why that matters.

## By use case

| Use case | Where to find it |
| --- | --- |
| Basic two-peer connect | [`browser-vanilla/basic-connect.html`](./browser-vanilla/basic-connect.html), [`node/basic-connect.js`](./node/basic-connect.js), [`browser-demo/demo.html`](./browser-demo/demo.html) |
| Group chat / mesh (many peers, broadcast to all) | [`browser-vanilla/group-chat-mesh.html`](./browser-vanilla/group-chat-mesh.html), [`node/group-chat-mesh.js`](./node/group-chat-mesh.js) |
| 1:1 random matchmaking | [`react/Matchmaking.jsx`](./react/Matchmaking.jsx), [`browser-demo/demo.html`](./browser-demo/demo.html) |
| File transfer | [`browser-vanilla/file-transfer.html`](./browser-vanilla/file-transfer.html) |
| Video/voice call | [`browser-vanilla/video-call.html`](./browser-vanilla/video-call.html), [`react/VideoCall.jsx`](./react/VideoCall.jsx) |
| Reconnect and error handling reference | [`node/reconnect-and-errors.js`](./node/reconnect-and-errors.js) |

## By environment

### `browser-vanilla/`

Plain HTML and JavaScript. No build step, no `npm install`. PeerJS is loaded from a CDN with a single `<script>` tag. Open any file directly in a browser, or serve the folder with any static file server.

- `basic-connect.html`: the smallest complete example, two peers finding each other and exchanging text messages.
- `group-chat-mesh.html`: the same idea extended to any number of peers in one basket.
- `file-transfer.html`: sends a file directly between two browsers.
- `video-call.html`: a 1:1 video/voice call using `peer.call()`.

### `browser-demo/`

- `demo.html`: a single clickable page. Pick "Open lobby" or "Find a match" with a button, no code reading required to try it out. Open it in two tabs to see it work.

### `node/`

Command-line peers, using PeerJS's native (beta) Node.js support.

```bash
cd node
npm install
node basic-connect.js          # run in two terminals
node group-chat-mesh.js        # run in three or more terminals
node reconnect-and-errors.js   # a single resilient peer, see comments inline
```

PeerJS's Node.js, Bun, and Deno support is currently in beta. If you hit WebRTC-related errors specific to your platform, the browser examples are the more battle-tested option.

### `react/`

Components, meant to be dropped into an existing React app.

```bash
npm install peerjs
```

- `Matchmaking.jsx`: pairs the current user with exactly one other user, then stops polling PeerBasket once matched.
- `VideoCall.jsx`: a 1:1 video/voice call component with proper cleanup on unmount (stops the camera, ends the call, destroys the peer).

## A note on the reconnect/error-handling example

Every example above includes basic error handling inline (a failed `fetch`, an HTTP error from PeerBasket, a connection that never opens). `node/reconnect-and-errors.js` is the one place where that logic is the entire point of the file rather than a side detail. Read it if you want to understand exactly what can go wrong and why each check exists, even if you end up using a browser example for your actual project.