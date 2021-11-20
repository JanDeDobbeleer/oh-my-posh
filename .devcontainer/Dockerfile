# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.177.0/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version (use -bullseye variants on local arm64/Apple Silicon): 1, 1.16, 1.17, 1-bullseye, 1.16-bullseye, 1.17-bullseye, 1-buster, 1.16-buster, 1.17-buster
ARG VARIANT=1-bullseye
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Choice] Node.js version: none, lts/*, 16, 14, 12, 10
ARG NODE_VERSION="none"
RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# Download and register the Microsoft repository GPG keys
RUN wget https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb && \
    dpkg -i packages-microsoft-prod.deb && \
    rm -f packages-microsoft-prod.deb

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y --no-install-recommends \
        fish \
        powershell \
        tmux \
    && apt-get clean

# Deploy latest stable oh-my-posh as a starting point:
RUN pwsh -Command Install-Module oh-my-posh -Scope AllUsers -Force && \
    chmod 777 -R /usr/local/share/powershell
# If you recompile the prompt in the devcontainer you may overwrite the executables at:
# /usr/local/share/powershell/Modules/oh-my-posh/<version>/bin/

# Deploy symlink of binary under /usr/bin/oh-my-posh for convenience access from non-Powershell shells:
RUN export ARCH=`uname -m`;\
    if [[ "$ARCH" -eq "aarch64" ]]; then \
        export POSH_BIN="$(find /usr/local/share/powershell/Modules/oh-my-posh/*/bin/posh-linux-arm64)";\
    elif [[ "$ARCH" -eq "armv7l" ]]; then \
        export POSH_BIN="$(find /usr/local/share/powershell/Modules/oh-my-posh/*/bin/posh-linux-arm)";\
    else \
        export POSH_BIN="$(find /usr/local/share/powershell/Modules/oh-my-posh/*/bin/posh-linux-amd64)";\
    fi;\
    ln -sf $POSH_BIN /usr/bin/oh-my-posh
# NOTE: devcontainers are Linux-only at this time but when
# Windows or Darwin is supported someone will need to improve
# the code logic above.

# Deploy all latest oh-my-posh themes at devcontainer build time:
ARG USERNAME=vscode
ARG POSHTHEMES_ROOT=/home/${USERNAME}/.poshthemes
ENV POSHTHEMES_ROOT=${POSHTHEMES_ROOT}
RUN mkdir -m=777 -p ${POSHTHEMES_ROOT} && \
    wget https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/themes.zip -O ${POSHTHEMES_ROOT}/themes.zip && \
    unzip ${POSHTHEMES_ROOT}/themes.zip -d ${POSHTHEMES_ROOT} && \
    chmod 777 ${POSHTHEMES_ROOT}/*.json && \
    rm ${POSHTHEMES_ROOT}/themes.zip

# Can be used to override the devcontainer prompt default theme:
ARG DEFAULT_POSH_THEME=jandedobbeleer
ENV DEFAULT_POSH_THEME=${DEFAULT_POSH_THEME}

# Deploy oh-my-posh prompt to Powershell:
COPY Microsoft.PowerShell_profile.ps1 /home/${USERNAME}/.config/powershell/Microsoft.PowerShell_profile.ps1

# Deploy oh-my-posh prompt to Fish:
COPY config.fish /home/${USERNAME}/.config/fish/config.fish

# Everything runs as root during build time, so we want
# to make sure the vscode user can edit these paths too:
RUN chmod 777 -R /home/${USERNAME}/.config

# Override vscode's own Bash prompt with oh-my-posh:
RUN sed -i 's/^__bash_prompt$/#&/' /home/${USERNAME}/.bashrc && \
    echo "eval \"\$(oh-my-posh --init --shell bash --config ${POSHTHEMES_ROOT}/${DEFAULT_POSH_THEME}.omp.json)\"" >> /home/${USERNAME}/.bashrc

# Override vscode's own ZSH prompt with oh-my-posh:
RUN sed -i 's/^source \$ZSH/oh-my-zsh.sh$/# source \$ZSH/oh-my-zsh.sh&/' /home/${USERNAME}/.zshrc && \
    echo "eval \"\$(oh-my-posh --init --shell zsh --config ${POSHTHEMES_ROOT}/${DEFAULT_POSH_THEME}.omp.json)\"" >> /home/${USERNAME}/.zshrc

# Set container timezone:
ARG TZ="UTC"
RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime

# [Optional] Uncomment the next line to use go get to install anything else you need
# RUN go get -x github.com/JanDeDobbeleer/battery

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1
