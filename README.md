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
  snaggle [--copy | --in-place] [--recursive] DIRECTORY DESTINATION

Flags:
      --copy        Copy entire directory contents to /DESTINATION/full/source/path
  -h, --help        help for snaggle
      --in-place    Snag in place: only snag dependencies & interpreter
  -r, --recursive   Recurse subdirectories & snag everything
  -v, --verbose     Output to stdout and process sequentially for readability
      --version     version for snaggle


In the form "snaggle FILE DESTINATION":
  FILE and all dependencies will be snagged to DESTINATION.
  An error will be returned if FILE is not a valid ELF binary.

In the form "snaggle DIRECTORY DESTINATION":
  All valid ELF binaries in DIRECTORY, and all their dependencies, will be snagged to DESTINATION.

Snaggle will hardlink (or copy, see notes):
- Executables              -> DESTINATION/bin
- Dynamic libraries (*.so) -> DESTINATION/lib64

Notes:
- Follows symlinks
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

## Example usage

```Dockerfile
# For example to get a locked-down nginx webserver running in 4 simple steps...

# You can use any base distro you like. I find fedora generally does well staying up to date with latest versions.
FROM fedora:latest AS installer

# 1. Install nginx and tini (for pid1)
RUN dnf install -y \
  nginx \
  tini

# 2. Install snaggle
COPY --from=ghcr.io/musicalninjadad/snaggle:latest /snaggle /bin/

# 3. Build the runtime root filesystem
WORKDIR /runtime

    # snaggle tini & nginx
    RUN snaggle /bin/tini . \
     && snaggle /bin/nginx .

    # add our config and data
    COPY nginx.conf ./etc/nginx/
    COPY index.html ./data/www/

    # Create the few standard locations nginx needs
    RUN \
        # for temp files
        mkdir --parents ./var/lib/nginx/tmp && chmod 1777 ./var/lib/nginx/tmp \
        # for logs & link logs to stdout & stderr
     && mkdir --parents ./var/log/nginx && chmod 0777 ./var/log/nginx \
     && ln -sf /dev/stdout ./var/log/nginx/access.log \
     && ln -sf /dev/stderr ./var/log/nginx/error.log \
        # for the pid file
     && mkdir ./run && chmod 1777 ./run

# 4. copy the minimal root filesystem to a new empty layer
FROM scratch AS runtime

COPY --from=installer /runtime /

USER 1000
EXPOSE 8000
ENTRYPOINT [ "tini", "--", "nginx" ]
```

## Known limitations

- only handles dynamic binaries with `/lib64/ld_linux...so` as an interpreter, no interpreter and static binaries.
- does not handle binaries compiled with dependencies in a custom `RUNPATH` or `RPATH` ([#13](https://github.com/MusicalNinjaDad/snaggle/issues/13))

## Planned improvements

Future versions will:

- provide standard profiles for apps with non-linked dependencies such as SSL certs, locales, gconv etc.
- further improve performance by bailing early on an error ([#37](https://github.com/MusicalNinjaDad/snaggle/issues/37))

## Why Go?

Historically this started as a python script, but I had to learn Go at some point - and this seemed like a good one. Plus it's much easier to `ADD` and use a single statically linked binary than a script and supporting interpreter.

That old python script is here in the git repo as v0.0.1 in case you're interested.
