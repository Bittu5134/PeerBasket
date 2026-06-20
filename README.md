PeerBasket eliminates separate join, leave, and status routes. Everything happens via a single API route. When a client pings a basket (our term for lobbies), the server registers or refreshes their presence in the swarm

"PeerBasket is a public, zero-overhead room discovery engine for PeerJS developers. By replacing fragile WebSockets with a single consolidated HTTP checkpoint, it registers presence and synchronization in a single network bound—eliminating connection state overhead

PeerBasket is a public, zero-config room discovery API for PeerJS. A single API route is used to register you to any arbitary lobby and return all the other peers present in the same lobby as you.

I often need a discovery server for my p2p projects but its a hassle to set it up a new one for each of them and setting up a PeerJS server would require a setting up a decent server in itself, thats why I madw this project that streamlines this process for devs who just want to get things done without worrying too much about infrastructre. This porject is free and will remain that way.

Here is the rephrased version turned into a sharp, objective pith:

PeerBasket eliminates the infrastructure overhead of spinning up dedicated PeerJS instances for every P2P project. Built for developers who just want to deploy peer-to-peer networks without configuration hassles, it streamlines swarm discovery and intentionally avoids authentication by design. This service is completely free and will always stay that way.




---


Devlog #1: Side Quest Project

Hello! this is gonna be my little sidequest project.

This is a sub-project of my ThumbScore and GeoShuffle Projects. In both of them I wanted to add p2p/multiplayer Features, but I was lacking Client Discovery features, thats why I firts thought of building a simple server for myself, But then I had the idea of Building a common server for everone!

For this project I had to spend time learning and Reading through Docs on multiple topics, like..
- Golang Net/Http
- PeerJS server (which is what I didnt had the capacity to host myself)
- Redis Caching
- Linux Server hosting

Ngl the last part took the most hours cause I havent every worked on a server myself, so its a new experince and lots of trial and errors.

Now Back to how this works
Its very Simple Honestly

a user pings the `/basket/:id` route and registers their peerJS ID, the route returns a list of other peers in the same basket. also this ping acts as a heartbeat for peer connectiion, so if a peer hasnt pinged the server in the last 30 seconds, then its removed

very simple logic ^-^

now ill go and try to setup a server