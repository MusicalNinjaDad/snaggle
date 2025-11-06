# Snaggle - snag the minimal files needed for a simple app container

Had enough of every container you pull having a full OS available inside? Create your own minimal app container easily by snaggling the binary and linked libraries.

<!-- AUTO-GENERATED via go generate -->
```text
ninjacoder@f52ce3a5f188:/workspaces/snaggle/snaggle$ ./snaggle --help
Snag a copy of a binary and all its dependencies to DESTINATION/bin & DESTINATION/lib64

Snaggle is designed to help create minimal runtime containers from pre-existing installations.
It may work for other use cases and I'd be interested to hear about them at:
https://github.com/MusicalNinjaDad/snaggle

Usage:
  snaggle [--in-place] FILE DESTINATION
  snaggle [--in-place] [--recursive] DIRECTORY DESTINATION

Flags:
  -h, --help        help for snaggle
      --in-place    Snag in place: only snag dependencies & interpreter
  -r, --recursive   Recurse subdirectories & snag everything
  -v, --verbose     Output to stdout and process sequentially for readability


In the form "snaggle FILE DESTINATION":
  FILE and all dependencies will be snagged to DESTINATION.
  An error will be returned if FILE is not a valid ELF binary.

In the form "snaggle DIRECTORY DESTINATION":
  All valid ELF binaries in DIRECTORY, and all their dependencies, will be snagged to DESTINATION.

Snaggle will hardlink (or copy, see notes):
- Executables              -> DESTINATION/bin
- Dynamic libraries (*.so) -> DESTINATION/lib64

Notes:
- Hardlinks will be created if possible.
- A copy will be performed if hardlinking fails for one of the following reasons:
    FILE/DIRECTORY & DESTINATION are on different filesystems or
    the user does not have permission to hardlink (e.g.
      https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
- Copies will retain the original filemode
- Copies will attempt to retain the original ownership, although this will likely fail if running as non-root
- Running with --verbose will be slower, not only due to processing stdout, but also as each file will be processed
  sequentially to provide readable output. Running silently will process all files and dependencies in parallel.

Exit Codes:
  0: Success
  1: Error
  2: Invalid command
  3: Panic
```
<!-- END AUTO-GENERATED -->

## Installing snaggle

### In a container (easiest)

  ```docker
  COPY --from=ghcr.io/musicalninjadad/snaggle /snaggle /bin/
  ```

### Download

> Grab the latest release binary (& SHA) from GitHub [`MusicalNinjaDad/snaggle/releases`](https://github.com/MusicalNinjaDad/snaggle/releases)

### Go install

> Install with `go install https://github.com/MusicalNinjaDad/snaggle@latest`

## Planned improvements

Future versions will:

- provide a guaranteed error type from the snaggle library & make exit code handling more robust in the cli
- provide standard profiles for apps with non-linked dependencies such as SSL certs, locales, gconv etc.

## Why Go?

Historically this started as a python script, but I had to learn Go at some point - and this seemed like a good one. Plus it's much easier to `ADD` and use a single statically linked binary than a script and supporting interpreter.

That old python script is here in the git repo as v0.0.1 in case you're interested.
