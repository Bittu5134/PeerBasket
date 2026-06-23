# PeerBasket

PeerBasket is a free, open-source, zero-config peer discovery API for [PeerJS](https://peerjs.com/).

[PeerJS](https://peerjs.com/) lets two browsers open a direct peer-to-peer connection, but only once both sides already know each other's peer ID. It has no built-in way for two peers to find each other in the first place. PeerBasket solves that one problem. You send your peer ID to a basket, which is any string that both sides have agreed on in advance, and PeerBasket sends back the peer ID of everyone else who is currently registered in that same basket. There is no separate step to create a basket first. Sending a peer ID to a basket name is enough to create it.

A hosted instance runs at `peerbasket.bittu.dev`, so most people will not need to run their own. Docs for the hosted instance live at [peerbasket.bittu.dev](https://peerbasket.bittu.dev/). This README covers the same API and how to self-host it.

## Why does this exist?

PeerBasket exists so that you can build a peer-to-peer app without running your own signaling server. It intentionally skips authentication. See [Security Considerations](#security-considerations) below to understand what that trade-off means in practice. PeerBasket is free and open source. If it is useful to you, consider [supporting it on Patreon](https://www.patreon.com/cw/LazyBittu).

## API

### `POST /basket/<basket_id>?limit=100`

`limit` is optional and defaults to 100.

Send your PeerJS [peer ID](https://peerjs.com/client/api/peer#id-string-2) in the request body. PeerBasket registers that peer ID under the `basket_id` given in the URL, which can be any string you choose, and returns the peer IDs of everyone else who is currently registered under that same `basket_id`.

**Request body**

```json
{
  "peer_id": "peerjs-abc-123"
}
```

**Response**

```json
{
  "basket_id": "room-101",
  "peers": ["peerjs-id-001", "peerjs-id-002"],
  "total_peers": 2,
  "peers_returned": 2
}
```

### Notes

- The `peers` list in the response includes your own peer ID. Filter it out on the client side before connecting, so that you do not try to connect to yourself.
- A peer is automatically dropped from its basket if it has not posted to that basket in the last 30 seconds.
- "Active" means a recent heartbeat, not a connection that is open right now. A peer that crashed a few seconds ago can still appear in the list for up to 30 seconds. Treat the moment a PeerJS connection actually opens as the real liveness check, and make sure your code handles the case where it never opens.
- **Rate limit:** PeerBasket allows 20 requests per minute. If you poll no faster than every 10 seconds, you will stay comfortably inside that limit. Exceeding it returns an HTTP 429 response.
- Basket IDs cannot be reserved. Anyone who knows a basket ID can join that basket and see everyone else who is in it.
- Use a long, random, unguessable string as your `basket_id` for anything other than a throwaway demo, the same way you would treat an unlisted URL. A v4 UUID works well. [`crypto.randomUUID()`](https://developer.mozilla.org/en-US/docs/Web/API/Crypto/randomUUID) is built into every modern browser and needs no extra library.
- The value of `peers_returned` will be lower than `total_peers` whenever `limit` is smaller than the basket's actual size.

## Example

A full quickstart, including creating a peer, connecting to others, and polling on an interval, lives in the [`examples`](./examples) folder.

## Self-hosting

PeerBasket is a Go server backed by Redis.

**Requirements**

- Go 1.21+
- A reachable Redis instance

**Run it**

```bash
git clone https://github.com/Bittu5134/PeerBasket.git
cd PeerBasket
go run main.go
```

**Configuration**

Set these as environment variables, or in a `.env` file in the project root:

| Variable     | Default          | Description                  |
| ------------ | ---------------- | ----------------------------- |
| `PORT`       | `8080`           | Port the server listens on    |
| `REDIS_ADDR` | `localhost:6379` | Address of your Redis instance |

## Security considerations

PeerBasket does no authentication, by design. A basket ID is the *only* thing that stands between a basket and the public. Treat a basket ID like an unlisted URL rather than a password, and pick something unguessable for anything that is not a throwaway demo, as recommended above.

This also means that PeerBasket cannot vouch for who is behind any peer ID it returns. Anyone who knows your basket ID can register their own peer into it, and your app will see that peer as just another member of the basket. If you are building anything beyond a quick prototype, verify identity at the application layer instead of trusting the raw peer list on its own. Two simple ways to do that are a shared secret exchanged out of band, or a handshake performed once the PeerJS connection actually opens.

## License

See [LICENSE](./LICENSE) for details.

## Links

- [Hosted docs](https://peerbasket.bittu.dev/)
- [GitHub repo](https://github.com/Bittu5134/PeerBasket)
- [Discord](https://discord.gg/CZdNvKaNNr)
- [Patreon](https://www.patreon.com/cw/LazyBittu)
- Built by [Bittu](https://bittu.dev)