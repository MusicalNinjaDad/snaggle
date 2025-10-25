FROM snaggle_devcontainer AS base

USER root
COPY --from=snaggle /snaggle /bin/

WORKDIR /runtime
RUN snaggle /sbin/bash . \
 && snaggle /sbin/which .

FROM scratch AS runtime

USER 1000
COPY --from=base /runtime /
ENTRYPOINT [ "/bin/bash" ]
CMD [ "-c", "'which which'" ]
