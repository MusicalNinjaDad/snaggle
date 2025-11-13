FROM fedora:latest@sha256:aa7befe5cfd1f0e062728c16453cd1c479d4134c7b85eac00172f3025ab0d522 AS installer

RUN dnf install -y \
  nginx \
  tini

COPY --from=ghcr.io/musicalninjadad/snaggle:v1.1.0 /snaggle /bin/

WORKDIR /runtime
    RUN mkdir ./tmp && chmod 1777 ./tmp \
     && snaggle /bin/tini . \
     && snaggle /bin/nginx .

FROM scratch AS runtime

COPY --from=installer /runtime /
USER 1000
ENTRYPOINT [ "tini", "--", "nginx" ]
