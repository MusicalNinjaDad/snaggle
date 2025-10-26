FROM fedora:latest@sha256:aa7befe5cfd1f0e062728c16453cd1c479d4134c7b85eac00172f3025ab0d522

# ---
# Setup base system ...
# ---

COPY <<EOF /etc/dnf/dnf.conf
[main]
gpgcheck=True
installonly_limit=3
clean_requirements_on_remove=True
best=False
skip_if_unavailable=True
install_weak_deps=False
assumeyes=true
# # Uncomment to disable man-pages
# tsflags=nodocs
EOF

# Create the default user
ARG USERNAME=ninjacoder
ARG USER_UID=1000
ARG USER_GID=${USER_UID}
RUN groupadd --gid ${USER_GID} ${USERNAME} \
&& useradd --uid ${USER_UID} --gid ${USER_GID} -m ${USERNAME} \
&& echo ${USERNAME} ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/${USERNAME} \
&& chmod 0440 /etc/sudoers.d/${USERNAME}

# ---
# Install ...
# ---

# Man itself, basic manpages (and awk we need next)
RUN dnf install \
    awk \
    man \
    man-db \
    man-pages \
&& dnf update \
# man pages for all the stuff which is already installed
&& dnf reinstall --skip-unavailable $(dnf list --installed | awk '{print $1}') \
# basic development tools
&& dnf install \
    bash-completion \
    curl \
    gh \
    git \
    glibc-all-langpacks \
    just \
    ping \
    which

# pre-commit via uv
# uv needs something from either curl or ping to allow it to do dns resolution
ENV UV_PYTHON_BIN_DIR=/bin
ENV UV_PYTHON_INSTALL_DIR=/usr/share
ENV UV_TOOL_BIN_DIR=/bin
ENV UV_TOOL_DIR=/usr/share
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/
RUN uv tool install pre-commit

# Go
RUN dnf install \
    golang \
    golangci-lint \
    golang-x-tools-goimports

# ---
# Final setup steps
# ---

# Set the default user
USER ${USERNAME}
