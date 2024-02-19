import asyncio

from typing import  Optional
from grpclib.client import Channel

from .gen.chaparral.v1 import AccessServiceStub, CommitServiceStub


class Client(Channel):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.access_service = AccessServiceStub(self)
        self.access_service.metadata = {}
        self.commit_service = CommitServiceStub(self)
        self.commit_service.metadata = {}

    def set_bearer_token(self, token: str):
        auth = "Bearer" + token
        self.access_service.metadata["authorization"] = auth
        self.access_service.metadata["authorization"] = auth

    async def a_get_version(self, obj_id: str, root_id: str = "", version: int = 0):
        return await self.access_service.get_object_version(storage_root_id = root_id,
                                                            object_id = obj_id,
                                                            version=version)

    def get_version(self, obj_id: str, root_id: str = "", version: int = 0):
        loop = asyncio.get_event_loop()
        result = loop.run_until_complete(self.a_get_version(obj_id, root_id, version))
        return result
