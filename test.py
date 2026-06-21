import asyncio
import random
import httpx

BASE_URL = "http://[2a01:4f9:3a:276e::893]:8001/"
BASKET_ID = "consolidated-room"


async def peer_lifecycle(peer_id: str):
    """Simulates a single peer checking in periodically, then disappearing."""
    async with httpx.AsyncClient(base_url=BASE_URL) as client:
        # Define how long this peer will stay in the lobby (15 to 40 seconds)
        lifetime = random.randint(15, 40)
        interval = 10  # Send single POST check-in every 10 seconds
        elapsed = 0

        print(f"🚀 [Spawn] {peer_id} is entering the lobby loop.")

        while elapsed < lifetime:
            try:
                # The single POST endpoint handles both registration and fetching
                res = await client.post(
                    f"/basket/{BASKET_ID}?limit=50", 
                    json={"peer_id": peer_id}
                )
                
                if res.status_code == 200:
                    data = res.json()
                    # Optional: Print what this specific peer sees in the room right now
                    print(f"💓 [Heartbeat/Sync] {peer_id} updated. Room count seen: {data.get('total_peers')}")
                else:
                    print(f"❌ [Error] {peer_id} failed check-in: {res.text}")
                    break
            except Exception as e:
                print(f"⚠️ [Connection Error] {peer_id}: {e}")
                break

            await asyncio.sleep(interval)
            elapsed += interval

        # When the loop ends, the peer simply stops pinging.
        # The Go backend's 60-second TTL math will sweep them away automatically.
        print(f"👻 [Ghost] {peer_id} stopped pinging (fading from presence cache)...")


async def monitor_basket():
    """Infinitely monitors the basket every 3 seconds using a dummy spectator ID."""
    async with httpx.AsyncClient(base_url=BASE_URL) as client:
        while True:
            await asyncio.sleep(3)
            try:
                # The monitor also uses the POST route since it's the only endpoint available.
                # It checks in as 'system-monitor' to safely inspect the pool.
                res = await client.post(
                    f"/basket/{BASKET_ID}", 
                    json={"peer_id": "system-monitor"}
                )
                if res.status_code == 200:
                    data = res.json()
                    peers = data.get("peers", [])
                    print(f"\n📊 [ROOM STATUS] Active count: {data.get('total_peers')} | Returned: {data.get('peers_returned')} | Peers: {peers}\n")
            except Exception as e:
                print(f"📊 [MONITOR ERROR]: {e}")


async def spawn_peers_infinitely():
    """Infinitely spawns new random peers over time to create live network churn."""
    peer_counter = 1
    while True:
        peer_id = f"peerjs-{peer_counter:04d}"
        peer_counter += 1

        # Fire off the peer lifecycle task completely non-blocking
        asyncio.create_task(peer_lifecycle(peer_id))

        # Wait a random interval before spawning the next client context
        await asyncio.sleep(random.uniform(3, 8))


async def main():
    print(f"🔄 Starting Consolidated Single-Route Simulation for: '{BASKET_ID}'")
    print("Press Ctrl+C to stop.")
    
    # Run the continuous spawner and the monitor alongside each other indefinitely
    await asyncio.gather(
        spawn_peers_infinitely(),
        monitor_basket()
    )


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n🛑 Simulation stopped by user.")