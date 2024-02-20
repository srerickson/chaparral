from typing import List
import hashlib
from pathlib import Path


def digest_dir(dir: Path, alg: str = "sha512") -> dict[str, List[Path]]:

    def __err(err):
        raise err
    digests: dict[str, List[Path]] = {}
    for parent, _, files in dir.walk(on_error=__err):
        for child in files:
            with (parent / child).open(mode='rb') as f:
                d = hashlib.file_digest(f, alg).hexdigest()
                n = Path(f.name).relative_to(dir)
                if d not in digests:
                    digests[d] = []
                digests[d].append(n)
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
