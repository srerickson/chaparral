import asyncio
import aiofiles
import chaparral
import os

token = os.getenv("CHAPARRAL_TOKEN")
host = "http://127.0.0.1:8080"
obj = "new-object-01"
root = "main"

async def main():
    async with chaparral.Client(host, token) as client:
        version = await client.aget_version(root, obj)
        # manifest = await client.aget_manifest("new-object-01", root="main")
        for digest, info in version.state.items():
            async for chunk in client.aiter_bytes(root, obj, digest):
                for path in info.paths:
                    async with aiofiles.open(path,'wb') as f:
                        await f.write(chunk)
asyncio.run(main())
