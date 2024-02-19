import hashlib
import os
from pathlib import Path


def digest_dir(name: str, alg: str = "sha512") -> dict[str,str]:
    digests = {}
    base = Path(name)
    def __err(err):
        raise err
    for parent, _, files in Path(name).walk(on_error=__err):
        for child in files:
            with (parent / child).open(mode='rb') as f:
                digest = hashlib.file_digest(f, alg)
                _name = Path(f.name).relative_to(name).as_posix()
                digests[_name] = digest.hexdigest()
    return digests

# async def process_files(directory):
#     tasks = []
#     async for dirpath, _, filenames in os.walk(directory):
#         for filename in filenames:
#             file_path = os.path.join(dirpath, filename)
#             tasks.append(calculate_sha256(file_path))
#     return await asyncio.gather(*tasks)

# async def main(directory):
#     results = await process_files(directory)
#     return results

# def run_asyncio_task(task):
#     loop = asyncio.get_event_loop()
#     return loop.run_until_complete(task)

# if __name__ == "__main__":
#     directory_path = "/path/to/your/directory"
#     results = run_asyncio_task(main(directory_path))

#     for file_path, sha256_checksum in results:
#         print(f"{file_path}: {sha256_checksum}")
