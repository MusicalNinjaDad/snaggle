# Snaggle - snag the minimal files needed for a simple app container

Had enough of every container you pull having a full OS available inside? Create your own minimal app container easily by snaggling the binary and linked libraries.

```text
ninjacoder@5747a297e3a1:/workspaces/snaggle/snaggle$ ./snaggle --help

Snag a copy of FILE and all its dependencies to DESTINATION/bin & DESTINATION/lib64

Snaggle is designed to help create minimal runtime containers from pre-existing installations.
It may work for other use cases and I'd be interested to hear about them at:
https://github.com/MusicalNinjaDad/snaggle

Usage:
  snaggle FILE DESTINATION [flags]

Flags:
  -h, --help   help for snaggle


Snaggle will hardlink (or copy, see notes):
- FILE -> DESTINATION/bin
- All dynamically linked dependencies -> DESTINATION/lib64

Note:
- Future versions intend to provide improved heuristics for destination paths, currently calling
  Snaggle(path/to/a.library.so) will place a.library.so in root/bin and you need to move it manually
- Hardlinks will be created if possible.
- A copy will be performed if hardlinking fails for one of the following reasons:
  - path & root are on different filesystems
  - the user does not have permission to hardlink (e.g.
    https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
- Copies will retain the original filemode
- Copies will attempt to retain the original ownership, although this will likely fail if running as non-root

Exit Codes:
  0: Success
  1: Error
  2: Invalid command
  3: Panic
```

## Installing snaggle

1. In a container:

    ```docker
    COPY --from ghcr.io/MusicalNinjaDad/snaggle /snaggle /bin/
    ```

1. Grab the the latest release binary from github [MusicalNinjaDad/snaggle](https://github.com/MusicalNinjaDad/snaggle)

1. Install with `go install https://github.com/MusicalNinjaDad/snaggle@latest`

## Why Go?

Historically this started as a python script, but I had to learn Go at some point - and this seemed like a good one. Plus it's much easier to `ADD` and use a single statically linked binary than a script and supporting interpreter.

That old python script is here in the git repo as v0.0.1 in case you're interested.
