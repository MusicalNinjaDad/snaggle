#!/usr/bin/env -S uv run --no-project --script
#
# /// script
# requires-python = ">=3.10"
# dependencies = ["pylddwrap"]
# ///

from pathlib import Path

import lddwrap
import sys
from contextlib import suppress

if __name__ == "__main__":
    try:
        binary = Path(sys.argv[1])
        destination = Path(sys.argv[2])
    except IndexError:
        msg = "Usage: create_container.py <binary> <destination>\n"
        msg += f"Only got {sys.argv}\n"
        sys.stderr.write(msg)
        sys.exit(1)

    bin_dir = destination / "bin"
    bin_dir.mkdir(parents=True, exist_ok=True)
    target = bin_dir / binary.name

    print(f"Linking {binary} -> {target}")
    target.hardlink_to(binary.resolve(strict=True))

    deps = lddwrap.list_dependencies(path=binary)
    for dep in deps:
        if library := dep.path:
            target = destination / library.relative_to("/")
            lib_dir = target.parent
            lib_dir.mkdir(parents=True, exist_ok=True)

            print(f"Linking {library} -> {target}")
            with suppress(FileExistsError):
                target.hardlink_to(library.resolve(strict=True))
