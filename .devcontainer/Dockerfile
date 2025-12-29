# See here for image contents: https://github.com/devcontainers/images/blob/main/src/go/.devcontainer/Dockerfile

# [Choice] Go version: 1, 1.24, 1.25, 1-trixie, 1.24-trixie, 1.25-trixie, 1-bookworm, 1.24-bookworm, 1.25-bookworm, 1-bullseye, 1.24-bullseye, 1.25-bullseye
ARG VARIANT=1-trixie
FROM mcr.microsoft.com/vscode/devcontainers/go:${VARIANT}

# [Choice] Node.js version: none, lts/*, 24, 22, 20
ARG NODE_VERSION="none"
RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# Install powershell
ARG PS_VERSION="7.5.4"
# powershell-7.5.4-linux-x64.tar.gz
# powershell-7.5.4-linux-arm64.tar.gz
RUN ARCH="$(dpkg --print-architecture)"; \
    if [ "${ARCH}" = "amd64" ]; then \
        PS_BIN="v$PS_VERSION/powershell-$PS_VERSION-linux-x64.tar.gz"; \
    elif [ "${ARCH}" = "arm64" ]; then \
        PS_BIN="v$PS_VERSION/powershell-$PS_VERSION-linux-arm64.tar.gz"; \
    elif [ "${ARCH}" = "armhf" ]; then \
        PS_BIN="v$PS_VERSION/powershell-$PS_VERSION-linux-arm32.tar.gz"; \
    fi; \
    wget https://github.com/PowerShell/PowerShell/releases/download/$PS_BIN -O pwsh.tar.gz; \
    mkdir /usr/local/pwsh && \
    tar Cxvfz /usr/local/pwsh pwsh.tar.gz && \
    rm pwsh.tar.gz && \
    chmod +x /usr/local/pwsh/pwsh

ENV PATH=$PATH:/usr/local/pwsh

RUN echo 'deb http://download.opensuse.org/repositories/shells:/fish:/release:/4/Debian_13/ /' | tee /etc/apt/sources.list.d/shells:fish:release:4.list; \
    curl -fsSL https://download.opensuse.org/repositories/shells:fish:release:4/Debian_13/Release.key | gpg --dearmor | tee /etc/apt/trusted.gpg.d/shells_fish_release_4.gpg > /dev/null; \
    apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y --no-install-recommends \
    fish \
    tmux \
    fzf \
    && apt-get clean

ARG USERNAME=vscode

# NOTE: devcontainers are Linux-only at this time but when
# Windows or Darwin is supported someone will need to improve
# the code logic above.

# Setup a neat little PowerShell experience
RUN pwsh -Command Install-Module posh-git -Scope AllUsers -Force; \
    pwsh -Command Install-Module z -Scope AllUsers -Force; \
    pwsh -Command Install-Module PSFzf -Scope AllUsers -Force; \
    pwsh -Command Install-Module Terminal-Icons -Scope AllUsers -Force;

# add the oh-my-posh path to the PATH variable
ENV PATH="$PATH:/home/${USERNAME}/bin"

# Deploy oh-my-posh prompt to Powershell:
COPY Microsoft.PowerShell_profile.ps1 /home/${USERNAME}/.config/powershell/Microsoft.PowerShell_profile.ps1

# Deploy oh-my-posh prompt to Fish:
COPY config.fish /home/${USERNAME}/.config/fish/config.fish

# Everything runs as root during build time, so we want
# to make sure the vscode user can edit these paths too:
RUN chmod 777 -R /home/${USERNAME}/.config

# Override vscode's own Bash prompt with oh-my-posh:
RUN sed -i 's/^__bash_prompt$/#&/' /home/${USERNAME}/.bashrc && \
    echo "eval \"\$(oh-my-posh init bash)\"" >> /home/${USERNAME}/.bashrc

# Override vscode's own ZSH prompt with oh-my-posh:
RUN echo "eval \"\$(oh-my-posh init zsh)\"" >> /home/${USERNAME}/.zshrc

# Set container timezone:
ARG TZ="UTC"
RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime

# [Optional] Uncomment the next line to use go get to install anything else you need
# RUN go get -x github.com/JanDeDobbeleer/battery

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1
