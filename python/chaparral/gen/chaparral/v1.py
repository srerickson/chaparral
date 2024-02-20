# Generated by the protocol buffer compiler.  DO NOT EDIT!
# sources: chaparral/v1/core.proto, chaparral/v1/access_service.proto, chaparral/v1/commit_service.proto
# plugin: python-betterproto
from dataclasses import dataclass
from datetime import datetime
from typing import Dict, List, Optional

import betterproto
import grpclib


@dataclass
class User(betterproto.Message):
    """user info for version block in inventory"""

    name: str = betterproto.string_field(1)
    address: str = betterproto.string_field(2)


@dataclass
class GetObjectVersionRequest(betterproto.Message):
    """
    GetObjectVersionRequest is used to request information about an object's
    state.
    """

    # The storage root id for the object to access. If not set, the default
    # storage root is used.
    storage_root_id: str = betterproto.string_field(1)
    # The object id to access (required).
    object_id: str = betterproto.string_field(2)
    # The version index for the object state. The default value is 0, which
    # refers to the most recent version.
    version: int = betterproto.int32_field(3)


@dataclass
class GetObjectVersionResponse(betterproto.Message):
    """
    GetObjectVersionResponse represents state for a specific object version.
    """

    # The object's storage root id. The empty string corresponds to the default
    # storage root.
    storage_root_id: str = betterproto.string_field(1)
    # The object's id
    object_id: str = betterproto.string_field(2)
    # The index for the object version represented by the state.
    version: int = betterproto.int32_field(3)
    # The object's most recent version index.
    head: int = betterproto.int32_field(4)
    # The object's digest algorithm (sha512 or sha256)
    digest_algorithm: str = betterproto.string_field(5)
    # The object's logical state represented as a map from digests to  file info.
    # Path entries in the file info represent logical filenames for the object
    # version state.
    state: Dict[str, "FileInfo"] = betterproto.map_field(
        6, betterproto.TYPE_STRING, betterproto.TYPE_MESSAGE
    )
    # The message associated with the object version
    message: str = betterproto.string_field(7)
    # The user information associated with the object version
    user: "User" = betterproto.message_field(8)
    # The timestamp associated witht he object version
    created: datetime = betterproto.message_field(9)
    # The OCFL specification version for the object version.
    spec: str = betterproto.string_field(10)


@dataclass
class GetObjectManifestRequest(betterproto.Message):
    """
    GetObjectManifestRequest is used to request details about all content files
    in an object
    """

    # The storage root id for the object to access. If not set, the default
    # storage root is used.
    storage_root_id: str = betterproto.string_field(1)
    # The object id to access (required).
    object_id: str = betterproto.string_field(2)


@dataclass
class GetObjectManifestResponse(betterproto.Message):
    """
    GetObjectManifestResponse represents all content files stored  in an object
    across all versions
    """

    # The storage root id for the object
    storage_root_id: str = betterproto.string_field(1)
    # The object id for the manifest
    object_id: str = betterproto.string_field(2)
    # The object's path relative to the OCFL Storage Root that contains it
    path: str = betterproto.string_field(3)
    # digest algorithm used for manifest keys
    digest_algorithm: str = betterproto.string_field(4)
    # manifest is a map of digest values to file info. Path entries represent
    # content paths relative to root of the OCFL object
    manifest: Dict[str, "FileInfo"] = betterproto.map_field(
        5, betterproto.TYPE_STRING, betterproto.TYPE_MESSAGE
    )
    # The OCFL specification version for the object
    spec: str = betterproto.string_field(6)


@dataclass
class FileInfo(betterproto.Message):
    # file size
    size: int = betterproto.int64_field(1)
    # one or more file paths for the content
    paths: List[str] = betterproto.string_field(2)
    # map of alternate digests alg -> digest
    fixity: Dict[str, str] = betterproto.map_field(
        3, betterproto.TYPE_STRING, betterproto.TYPE_STRING
    )


@dataclass
class CommitRequest(betterproto.Message):
    # storage_root_id is the id of the storage root for the object to
    # create/update
    storage_root_id: str = betterproto.string_field(1)
    # object_id is the id for the object to create/update
    object_id: str = betterproto.string_field(2)
    # version is used to set the expected number for the new version. If set to
    # 0, the HEAD+1 is assumed.
    version: int = betterproto.int32_field(3)
    # User information for the commit. The user name and email are saved with the
    # new object version.
    user: "User" = betterproto.message_field(4)
    # Commit message. The message is saved with the new object version.
    message: str = betterproto.string_field(5)
    # state is a map of paths to digests using digest_algorithm
    state: Dict[str, str] = betterproto.map_field(
        6, betterproto.TYPE_STRING, betterproto.TYPE_STRING
    )
    # digest_algorithm is the id for the digest algorithm used in state. It  must
    # be 'sha512' or 'sha256'
    digest_algorithm: str = betterproto.string_field(7)
    content_sources: List["CommitRequestContentSourceItem"] = betterproto.message_field(
        8
    )


@dataclass
class CommitRequestContentSourceItem(betterproto.Message):
    # get new content from the uploader
    uploader: "CommitRequestUploaderSource" = betterproto.message_field(1, group="item")
    # get new content from an existing object
    object: "CommitRequestObjectSource" = betterproto.message_field(2, group="item")


@dataclass
class CommitRequestObjectSource(betterproto.Message):
    storage_root_id: str = betterproto.string_field(1)
    object_id: str = betterproto.string_field(2)


@dataclass
class CommitRequestUploaderSource(betterproto.Message):
    uploader_id: str = betterproto.string_field(1)


@dataclass
class CommitResponse(betterproto.Message):
    pass


@dataclass
class DeleteObjectRequest(betterproto.Message):
    """DeleteObjectRequest is used to delete an object and its files."""

    storage_root_id: str = betterproto.string_field(1)
    object_id: str = betterproto.string_field(2)


@dataclass
class DeleteObjectResponse(betterproto.Message):
    pass


@dataclass
class NewUploaderRequest(betterproto.Message):
    """
    NewUploaderRequest is used to create new uploaders where files can be
    uploaded.
    """

    # a list of digest algorithms use to digest files uploaded to the uploader.
    # The list must include `sha512` or `sha256`.
    digest_algorithms: List[str] = betterproto.string_field(1)
    # An optional uploader description
    description: str = betterproto.string_field(2)


@dataclass
class NewUploaderResponse(betterproto.Message):
    uploader_id: str = betterproto.string_field(1)
    # algorithm used to digest uploaded data
    digest_algorithms: List[str] = betterproto.string_field(2)
    # optional description (may be empty)
    description: str = betterproto.string_field(3)
    # ID for user who created uploader (may be empty)
    user_id: str = betterproto.string_field(4)
    # timestamp when uploader was created
    created: datetime = betterproto.message_field(5)
    # path for uploading content to the uploader
    upload_path: str = betterproto.string_field(6)


@dataclass
class GetUploaderRequest(betterproto.Message):
    """
    GetUploaderRequest is used to access information about an existing uploader
    """

    uploader_id: str = betterproto.string_field(1)


@dataclass
class GetUploaderResponse(betterproto.Message):
    """GetUploadResponse represent information about an uploader"""

    # uploader's unique ID
    uploader_id: str = betterproto.string_field(1)
    # algorithm used to digest uploaded data
    digest_algorithms: List[str] = betterproto.string_field(2)
    # optional description (may be empty)
    description: str = betterproto.string_field(3)
    # ID for user who created uploader (may be empty)
    user_id: str = betterproto.string_field(4)
    # timestamp when uploader was created
    created: datetime = betterproto.message_field(5)
    # path for uploading content to the uploader
    upload_path: str = betterproto.string_field(6)
    # list of uploads in the uploader
    uploads: List["GetUploaderResponseUpload"] = betterproto.message_field(7)


@dataclass
class GetUploaderResponseUpload(betterproto.Message):
    # map of algorithm name to digest value for the upload
    digests: Dict[str, str] = betterproto.map_field(
        1, betterproto.TYPE_STRING, betterproto.TYPE_STRING
    )
    # size of the upload in bytes
    size: int = betterproto.int64_field(2)


@dataclass
class ListUploadersRequest(betterproto.Message):
    """ListUploaderRequest is used to access a list of uploaders."""

    pass


@dataclass
class ListUploadersResponse(betterproto.Message):
    """ListUploaderResponse includes a list of uploaders"""

    uploaders: List["ListUploadersResponseItem"] = betterproto.message_field(1)


@dataclass
class ListUploadersResponseItem(betterproto.Message):
    uploader_id: str = betterproto.string_field(1)
    # creation date for the uploader
    created: datetime = betterproto.message_field(2)
    # optional description (may be empty)
    description: str = betterproto.string_field(3)
    # user id for the uploader (may be empty)
    user_id: str = betterproto.string_field(4)


@dataclass
class DeleteUploaderRequest(betterproto.Message):
    """
    DeleteUploaderRequest is used to delete an uploader and all its files.
    """

    uploader_id: str = betterproto.string_field(1)


@dataclass
class DeleteUploaderResponse(betterproto.Message):
    pass


class AccessServiceStub(betterproto.ServiceStub):
    """AccessService provides endpoints for reading OCFL objects."""

    async def get_object_version(
        self, *, storage_root_id: str = "", object_id: str = "", version: int = 0
    ) -> GetObjectVersionResponse:
        """
        GetObjectVersion returns details about the logical state of an OCFL
        object version.
        """

        request = GetObjectVersionRequest()
        request.storage_root_id = storage_root_id
        request.object_id = object_id
        request.version = version

        return await self._unary_unary(
            "/chaparral.v1.AccessService/GetObjectVersion",
            request,
            GetObjectVersionResponse,
        )

    async def get_object_manifest(
        self, *, storage_root_id: str = "", object_id: str = ""
    ) -> GetObjectManifestResponse:
        """
        GetObjectManifest returns digests, sizes, and fixity information for
        all content associated with an object across all its versions.
        """

        request = GetObjectManifestRequest()
        request.storage_root_id = storage_root_id
        request.object_id = object_id

        return await self._unary_unary(
            "/chaparral.v1.AccessService/GetObjectManifest",
            request,
            GetObjectManifestResponse,
        )


class CommitServiceStub(betterproto.ServiceStub):
    """
    CommitService provides an API for creating, updating, and deleting OCFL
    objects.
    """

    async def commit(
        self,
        *,
        storage_root_id: str = "",
        object_id: str = "",
        version: int = 0,
        user: Optional["User"] = None,
        message: str = "",
        state: Optional[Dict[str, str]] = None,
        digest_algorithm: str = "",
        content_sources: List["CommitRequestContentSourceItem"] = [],
    ) -> CommitResponse:
        """Commit creates or updates individual OCFL objects"""

        request = CommitRequest()
        request.storage_root_id = storage_root_id
        request.object_id = object_id
        request.version = version
        if user is not None:
            request.user = user
        request.message = message
        request.state = state
        request.digest_algorithm = digest_algorithm
        if content_sources is not None:
            request.content_sources = content_sources

        return await self._unary_unary(
            "/chaparral.v1.CommitService/Commit",
            request,
            CommitResponse,
        )

    async def new_uploader(
        self, *, digest_algorithms: List[str] = [], description: str = ""
    ) -> NewUploaderResponse:
        """
        NewUploader creates a new uploader where content can be uploaded before
        committing it to an object..
        """

        request = NewUploaderRequest()
        request.digest_algorithms = digest_algorithms
        request.description = description

        return await self._unary_unary(
            "/chaparral.v1.CommitService/NewUploader",
            request,
            NewUploaderResponse,
        )

    async def get_uploader(self, *, uploader_id: str = "") -> GetUploaderResponse:
        """GetUploader returns details for a specific uploader"""

        request = GetUploaderRequest()
        request.uploader_id = uploader_id

        return await self._unary_unary(
            "/chaparral.v1.CommitService/GetUploader",
            request,
            GetUploaderResponse,
        )

    async def list_uploaders(self) -> ListUploadersResponse:
        """ListUploaders returns a list of uploaders."""

        request = ListUploadersRequest()

        return await self._unary_unary(
            "/chaparral.v1.CommitService/ListUploaders",
            request,
            ListUploadersResponse,
        )

    async def delete_uploader(self, *, uploader_id: str = "") -> DeleteUploaderResponse:
        """DeleteUploader deletes an uploader and files uploaded to it."""

        request = DeleteUploaderRequest()
        request.uploader_id = uploader_id

        return await self._unary_unary(
            "/chaparral.v1.CommitService/DeleteUploader",
            request,
            DeleteUploaderResponse,
        )

    async def delete_object(
        self, *, storage_root_id: str = "", object_id: str = ""
    ) -> DeleteObjectResponse:
        """DeleteObject permanently deletes an existing OCFL object."""

        request = DeleteObjectRequest()
        request.storage_root_id = storage_root_id
        request.object_id = object_id

        return await self._unary_unary(
            "/chaparral.v1.CommitService/DeleteObject",
            request,
            DeleteObjectResponse,
        )