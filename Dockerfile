FROM snaggle_devcontainer AS builder
USER 1000
WORKDIR /workspaces/snaggle
COPY --chown=1000:1000 . .
RUN go test -v ./...

WORKDIR /workspaces/snaggle/cmd/snaggle
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

FROM scratch AS runtime
COPY --from=builder --chown=root:root /workspaces/snaggle/cmd/snaggle/snaggle /
ENTRYPOINT [ "/snaggle" ]
CMD [ "--help" ]
