import asyncio
import chaparral
import os

token = os.getenv("CHAPARRAL_TOKEN")
host = "http://127.0.0.1:8080"


async def main():
    async with chaparral.Client(host, token) as client:
        result = await client.a_get_version("new-object-01")
        print(result)
asyncio.run(main())
