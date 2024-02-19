from typing import Optional
from httpx import AsyncClient


# routes
access_svc_rt = "chaparral.v1.AccessService"
get_version_rt = access_svc_rt + "/GetObjectVersion"


class Client(AsyncClient):
    def __init__(self,
                 base_url: str = "http://127.0.0.1:8080",
                 token: Optional[str] = None):
        headers = {
            "content-type": "application/json",
            "authorization": "Bearer " + token
        }
        super().__init__(base_url=base_url,
                         http2=True,
                         headers=headers)

    async def a_get_version(self, obj: str, root: str = "", ver: int = 0):
        body = {
            "storageRootId": root,
            "objectId": obj,
            "version": ver
        }
        return await self.post(get_version_rt, json=body)

    # def get_version(self, obj: str, root: str = "", ver: int = 0):
    #     loop = asyncio.get_event_loop()
    #     result = loop.run_until_complete(self.a_get_version(obj, root, ver))
    #     return result
