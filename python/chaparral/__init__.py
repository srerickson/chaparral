from typing import Optional, List
from httpx import AsyncClient, Response
from pydantic import BaseModel, Field


# routes
_access_svc_rt = "chaparral.v1.AccessService"
_get_manifest_rt = _access_svc_rt + "/GetObjectManifest"
_get_version_rt = _access_svc_rt + "/GetObjectVersion"
_get_blob_rt = _access_svc_rt + "/download"

class FileInfo(BaseModel):
    paths: List[str]
    size: int = Field(default=0)
    fixity: dict[str,str] = Field(default={})
    

class User(BaseModel):
    name: str
    address: str = Field(default="")


class ObjectManifest(BaseModel):
    storage_root_id: str = Field(default="", alias="storageRootId")
    object_id: str = Field(alias="objectId")
    path: str
    spec: str
    manifest: dict[str, FileInfo]


class ObjectVersion(BaseModel):
    storage_root_id: str = Field(default="", alias="storageRootId")
    object_id: str = Field(alias="objectId")
    version: int
    head: int
    digest_algorithm: str = Field(alias="digestAlgorithm")
    spec: str
    user: Optional[User] = Field(default=None)
    message: str = Field(default="")
    state: dict[str, FileInfo]


class Client(AsyncClient):
    def __init__(self,
                 base_url: str = "http://127.0.0.1:8080",
                 token: Optional[str] = None):
        headers = {
            "content-type": "application/json",
        }
        if token is not None:
            headers["authorization"] = "Bearer " + token
        super().__init__(base_url=base_url,
                         http2=True,
                         headers=headers)

    async def aget_manifest(self, root: str, obj: str) -> ObjectManifest:
        body = {
            "storageRootId": root,
            "objectId": obj,
        }
        result = await self.post(_get_manifest_rt, json=body)
        result.raise_for_status()
        return ObjectManifest(**result.json())

    async def aget_version(self, root: str, obj: str, ver: int = 0) -> ObjectVersion:
        body = {
            "storageRootId": root,
            "objectId": obj,
            "version": ver
        }
        result = await self.post(_get_version_rt, json=body)
        result.raise_for_status()
        return ObjectVersion(**result.json())

    async def aiter_bytes(self,
                        root: str,
                        obj: str,
                        digest: str,
                        chunk_size: Optional[int] = None):
        params = {
            "storage_root": root,
            "object_id": obj,
            "digest": digest
        }
        async with self.stream('GET', _get_blob_rt, params=params) as response:
            response.raise_for_status()
            async for chunk in response.aiter_bytes(chunk_size):
                yield chunk
            print("done")

    # def get_version(self, obj: str, root: str = "", ver: int = 0):
    #     loop = asyncio.get_event_loop()
    #     result = loop.run_until_complete(self.a_get_version(obj, root, ver))
    #     return result
