#!/usr/bin/env sh

install_dir=""

error() {
    printf "\x1b[31m$1\e[0m\n"
    exit 1
}

info() {
    printf "$1\n"
}

warn() {
    printf "⚠️  \x1b[33m$1\e[0m\n"
}

help() {
    # Display Help
    echo "Installs Oh My Posh"
    echo
    echo "Syntax: install.sh [-h|d]"
    echo "options:"
    echo "h     Print this Help."
    echo "d     Specify the installation directory. Defaults to /usr/local/bin or the directory where oh-my-posh is installed."
    echo
}

while getopts ":hd:" option; do
   case $option in
      h) # display Help
         help
         exit;;
      d) # Enter a name
         install_dir=$OPTARG;;
     \?) # Invalid option
         echo "Invalid option command line option. Use -h for help."
         exit 1
   esac
done

SUPPORTED_TARGETS="linux-amd64 linux-arm linux-arm64 darwin-amd64 darwin-arm64"

validate_dependency() {
    if ! command -v $1 >/dev/null; then
        error "$1 is required to install Oh My Posh. Please install $1 and try again.\n"
    fi
}

validate_dependencies() {
    validate_dependency curl
    validate_dependency unzip
    validate_dependency realpath
    validate_dependency dirname
}

set_install_directory() {
    if [ -n "$install_dir" ]; then
        # expand directory
        install_dir="${install_dir/#\~/$HOME}"
        return 0
    fi

    # check if we have oh-my-posh installed, if so, use the executable directory
    # to install into and follow symlinks
    if command -v oh-my-posh >/dev/null; then
        posh_dir=$(command -v oh-my-posh)
        real_dir=$(realpath $posh_dir)
        install_dir=$(dirname $real_dir)
        info "Oh My Posh is already installed, updating existing installation in:"
        info "  ${install_dir}"
    else
        install_dir="/usr/local/bin"
    fi
}

validate_install_directory() {
    if [ ! -d "$install_dir" ]; then
        error "Directory ${install_dir} does not exist, set a different directory and try again."
    fi

    # check if we can write to the install directory
    if [ ! -w $install_dir ]; then
        error "Cannot write to ${install_dir}. Please run the script with sudo and try again:\n  sudo ./install.sh"
    fi

    # check if the directory is in the PATH
    good=$(
        IFS=:
        for path in $PATH; do
        if [ "${path%/}" = "${install_dir}" ]; then
            printf 1
            break
        fi
        done
    )

    if [ "${good}" != "1" ]; then
        warn "Installation directory ${install_dir} is not in your \$PATH"
    fi
}

install() {
    arch=$(detect_arch)
    platform=$(detect_platform)
    target="${platform}-${arch}"

    good=$(
        IFS=" "
        for t in $SUPPORTED_TARGETS; do
        if [ "${t}" = "${target}" ]; then
            printf 1
            break
        fi
        done
    )

    if [ "${good}" != "1" ]; then
        error "${arch} builds for ${platform} are not available for Oh My Posh"
    fi

    info "\nℹ️  Installing oh-my-posh for ${target} in ${install_dir}"

    executable=${install_dir}/oh-my-posh
    url=https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/posh-${target}

    info "⬇️  Downloading oh-my-posh from ${url}"

    curl -s -L https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/posh-${target} -o $executable
    chmod +x $executable

    # install themes in cache
    cache_dir=$(oh-my-posh cache path)

    info "🎨 Installing oh-my-posh themes in ${cache_dir}/themes\n"

    theme_dir="${cache_dir}/themes"
    mkdir -p $theme_dir
    curl -s -L https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/themes.zip -o ${cache_dir}/themes.zip
    unzip -o -q ${cache_dir}/themes.zip -d $theme_dir
    chmod u+rw ${theme_dir}/*.omp.*
    rm ${cache_dir}/themes.zip

    info "🚀 Installation complete.\n\nYou can follow the instructions at https://ohmyposh.dev/docs/installation/prompt"
    info "to setup your shell to use oh-my-posh.\n"
    info "If you want to use a built-in theme, you can find them in the ${theme_dir} directory:"
    info "  oh-my-posh init {shell} --config ${theme_dir}/{theme}.omp.json\n"
}

detect_arch() {
  arch="$(uname -m | tr '[:upper:]' '[:lower:]')"

  case "${arch}" in
    x86_64) arch="amd64" ;;
    armv*) arch="arm" ;;
    arm64) arch="arm64" ;;
  esac

  if [ "${arch}" = "arm64" ] && [ "$(getconf LONG_BIT)" -eq 32 ]; then
    arch=arm
  fi

  printf '%s' "${arch}"
}


detect_platform() {
  platform="$(uname -s | tr '[:upper:]' '[:lower:]')"

  case "${platform}" in
    linux) platform="linux" ;;
    darwin) platform="darwin" ;;
  esac

  printf '%s' "${platform}"
}

validate_dependencies
set_install_directory
validate_install_directory
install
