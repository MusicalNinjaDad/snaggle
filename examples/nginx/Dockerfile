# For example to get a locked-down nginx webserver running in 4 simple steps...

# You can use any base distro you like. I find fedora generally does well staying up to date with latest versions.
FROM fedora:latest@sha256:aa7befe5cfd1f0e062728c16453cd1c479d4134c7b85eac00172f3025ab0d522 AS installer

# 1. Install nginx and tini (for pid1)
RUN dnf install -y \
  nginx \
  tini

# 2. Install snaggle
COPY --from=ghcr.io/musicalninjadad/snaggle:v1.1.0 /snaggle /bin/

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
