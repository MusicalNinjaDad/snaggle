FROM fedora:latest@sha256:aa7befe5cfd1f0e062728c16453cd1c479d4134c7b85eac00172f3025ab0d522 AS installer

RUN dnf install -y \
  nginx \
  tini

COPY --from=ghcr.io/musicalninjadad/snaggle:v1.1.0 /snaggle /bin/

WORKDIR /runtime
    RUN snaggle /bin/tini . \
     && snaggle /bin/nginx .

    COPY nginx.conf ./etc/nginx/

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

FROM scratch AS runtime

COPY --from=installer /runtime /

USER 1000
EXPOSE 8000
ENTRYPOINT [ "tini", "--", "nginx" ]
