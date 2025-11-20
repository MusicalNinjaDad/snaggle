# 1. Set up a suitable build environment
FROM snaggle_devcontainer AS builder
USER root

#2. Build the app
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 go build -ldflags '-linkmode=external' -buildmode=pie .

# 3. Install snaggle
COPY --from=ghcr.io/musicalninjadad/snaggle:latest /snaggle /bin/

# 4. Build the runtime root filesystem
WORKDIR /runtime
RUN snaggle /build/hello .

# 5. copy the minimal root filesystem to a new empty layer
FROM scratch AS runtime
COPY --from=builder /runtime /

USER 1000
ENTRYPOINT [ "hello" ]
