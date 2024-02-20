import asyncio
import chaparral
import os

token = os.getenv("CHAPARRAL_TOKEN")
host = "https://api.chaparral.io"
obj = "index"
root = "restricted"


async def main():
    async with chaparral.Client(host, token) as client:
        await client.apull("tmp", root, obj)


asyncio.run(main())
