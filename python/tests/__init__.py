import asyncio
import chaparral
import os

token = os.getenv("CHAPARRAL_TOKEN")
host = "127.0.0.1"
port = "8080"

if token == "":
    print("warning: token isn't set")

async def main():
    async with chaparral.Client(host, port) as client:    
        client.set_bearer_token(token)
        result = await client.a_get_version("new-object-01")
        paths = [p for k, val in result.state.items() for p in val.paths ]
        print(paths)
asyncio.run(main())

