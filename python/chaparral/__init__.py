import asyncio

from typing import  Optional
from grpclib.client import Channel

from .gen.chaparral.v1 import AccessServiceStub, CommitServiceStub


class Client:
    def __init__(self, 
                 host: Optional[str] = "127.0.0.1",
                 port: Optional[str] = "8080",
                 tls = None,
                 token: Optional[str] = None):
        self.channel = Channel(host=host,port=port)
        self.headers = {}
        if token:
            self.headers["authorization"] = "Bearer "+ token
        self.access_service = AccessServiceStub(self.channel)
        self.access_service.metadata = self.headers
        self.commit_service = CommitServiceStub(self.channel)
        self.commit_service.metadata = self.headers

    def close(self):
        self.channel.close()

    async def a_get_version(self, obj_id: str, root_id: str = "", version: int = 0):
        return await self.access_service.get_object_version(storage_root_id = root_id,
                                                            object_id = obj_id,
                                                            version=version)

    def get_version(self, obj_id: str, root_id: str = "", version: int = 0):
        loop = asyncio.get_event_loop()
        result = loop.run_until_complete(self.a_get_version(obj_id, root_id, version))
        return result
