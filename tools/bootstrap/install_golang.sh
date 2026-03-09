#!/bin/bash
# shellcheck disable=SC2016
set -e

[ -z "$GOROOT" ] && GOROOT="$HOME/.go"
[ -z "$GOPATH" ] && GOPATH="$HOME/go"

# Function to detect the latest Go version from go.dev
get_latest_version() {
    local latest_version=""

    if hash wget 2>/dev/null; then
        latest_version=$(wget -qO- "https://go.dev/VERSION?m=text" 2>/dev/null | head -n 1 | sed 's/go//')
    elif hash curl 2>/dev/null; then
        latest_version=$(curl -sL "https://go.dev/VERSION?m=text" 2>/dev/null | head -n 1 | sed 's/go//')
    fi

    # Validate version format
    if [[ "$latest_version" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
        echo "$latest_version"
    else
        echo ""
    fi
}

OS="$(uname -s)"
ARCH="$(uname -m)"

case $OS in
    "Linux")
        case $ARCH in
        "x86_64")
            ARCH=amd64
            ;;
        "aarch64")
            ARCH=arm64
            ;;
        "armv6" | "armv7l")
            ARCH=armv6l
            ;;
        "armv8")
            ARCH=arm64
            ;;
        "i686")
            ARCH=386
            ;;
        .*386.*)
            ARCH=386
            ;;
        esac
        PLATFORM="linux-$ARCH"
    ;;
    "Darwin")
          case $ARCH in
          "x86_64")
              ARCH=amd64
              ;;
          "arm64")
              ARCH=arm64
              ;;
          esac
        PLATFORM="darwin-$ARCH"
    ;;
esac

print_help() {
    echo "Usage: bash goinstall.sh OPTIONS"
    echo -e "\nOPTIONS:"
    echo -e "  --version\tSpecify a version number to install"
}

if [ -z "$PLATFORM" ]; then
    echo "Your operating system is not supported by the script."
    exit 1
fi

if [ "$1" == "--help" ]; then
    print_help
    exit 0
elif [ "$1" == "--version" ]; then
    if [ -z "$2" ]; then # Check if --version has a second positional parameter
        echo "Please provide a version number for: $1"
    else
        VERSION=$2
    fi
elif [ ! -z "$1" ]; then
    echo "Unrecognized option: $1"
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo "You should specify a version number to install using --version."
    exit 1
fi

if [ -d "$GOROOT" ]; then
    echo "The Go install directory ($GOROOT) already exists. Exiting."
    exit 1
fi

PACKAGE_NAME="go$VERSION.$PLATFORM.tar.gz"
TEMP_DIRECTORY=$(mktemp -d)

echo "Downloading $PACKAGE_NAME ..."
if hash wget 2>/dev/null; then
    wget https://dl.google.com/go/$PACKAGE_NAME -O "$TEMP_DIRECTORY/go.tar.gz"
else
    curl -o "$TEMP_DIRECTORY/go.tar.gz" https://dl.google.com/go/$PACKAGE_NAME
fi

if [ $? -ne 0 ]; then
    echo "Download failed! Exiting."
    exit 1
fi

echo "Extracting File..."
mkdir -p "$GOROOT"
tar -C "$GOROOT" --strip-components=1 -xzf "$TEMP_DIRECTORY/go.tar.gz"
mkdir -p "${GOPATH}/"{src,pkg,bin}
echo -e "\nGo $VERSION was installed into $GOROOT."
rm -f "$TEMP_DIRECTORY/go.tar.gz"