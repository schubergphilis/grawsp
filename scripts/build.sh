#!/usr/bin/env bash

# ######################################################################################################################
# VARIABLES
# ######################################################################################################################

# General
export PROGRAM_NAME="grawsp"

# Directories
export BUILD_DIR="build"
export CMD_DIR="cmd"
export DIST_DIR="dist"

# Go
export CGO_ENABLED=1

# ######################################################################################################################
# FUNCTIONS
# ######################################################################################################################

_get_commit() {
    COMMIT=$(git rev-parse --short HEAD)

    if [ "$COMMIT" == "" ]; then
        echo "unknown"
    else
        echo $COMMIT
    fi
}

_get_version() {
    REV="unknown"

    if [ "$CI" == "" ]; then
        TAG_COMMIT=$(git rev-list --tags --max-count=1)

        if [ "$TAG_COMMIT" == "" ]; then
            REV=$(_get_commit)
        else
            REV=$(git describe --tags `git rev-list --tags --max-count=1`)
        fi
    else
        if [ "$GITHUB_SHA" != "" ]; then
            REV=$GITHUB_SHA
        fi
    fi

    echo $REV
}

_build() {
    BUILD_GOLANG_VERSION=$(go version | cut -d " " -f 3 | cut -c3-)
    BUILD_ARCH=$(echo $1 | cut -d/ -f2)
    BUILD_COMMIT=$(_get_commit)
    BUILD_OS=$(echo $1 | cut -d/ -f1)
    BUILD_TIME=$(date --rfc-3339=seconds)
    BUILD_VERSION=$(_get_version)
    BUILD_FILE="$BUILD_DIR/$BUILD_OS-$BUILD_ARCH/$PROGRAM_NAME"
    BUILD_LDFLAGS=""
    BUILD_MODE=$2

    if [ "$BUILD_MODE" == "release" ]; then
        BUILD_LDFLAGS="$BUILD_LDFLAGS -w -s"
        BUILD_LDFLAGS="$BUILD_LDFLAGS -X 'github.com/schubergphilis/grawsp/internal/meta.BuildTime=$BUILD_TIME'"
        BUILD_LDFLAGS="$BUILD_LDFLAGS -X 'github.com/schubergphilis/grawsp/internal/meta.Commit=$BUILD_COMMIT'"
        BUILD_LDFLAGS="$BUILD_LDFLAGS -X 'github.com/schubergphilis/grawsp/internal/meta.Version=$BUILD_VERSION'"
    fi

    echo "---------------------------------------------------------------------"
    echo "Building $PROGRAM_NAME for..."
    echo "---------------------------------------------------------------------"
    echo "Go:         $BUILD_GOLANG_VERSION"
    echo "Arch:       $BUILD_ARCH"
    echo "OS:         $BUILD_OS"
    echo "Timestamp:  $BUILD_TIME"
    echo "Commit:     $BUILD_COMMIT"
    echo "Version:    $BUILD_VERSION"
    echo "Flags:      $BUILD_LDFLAGS"
    echo "---------------------------------------------------------------------"
    echo ""

    export GOOS=$BUILD_OS
    export GOARCH=$BUILD_ARCH

    go build                             \
        -o "$BUILD_FILE"                 \
        -race                            \
        -mod=mod                         \
        -ldflags="$BUILD_LDFLAGS"        \
        "$CMD_DIR/$PROGRAM_NAME/main.go"

    if [ "$BUILD_MODE" == "release" ]; then
        TARBALL="$DIST_DIR/$PROGRAM_NAME-$BUILD_VERSION-$BUILD_OS-$BUILD_ARCH.tar.gz"

        echo "---------------------------------------------------------------------"
        echo "Releasing $TARBALL"
        echo "---------------------------------------------------------------------"

        tar -czvf "$TARBALL" -C $(dirname $BUILD_FILE) .

        echo "---------------------------------------------------------------------"
        echo ""
    fi
}

_usage() {
    echo "Usage: $0 [-r] -p <string>"
    echo ""
    echo "    -r  enable release mode"
    echo "    -p  platform, example: linux/amd64"
    echo ""
}

# ######################################################################################################################
# MAIN
# ######################################################################################################################

if [ ! -d ".git" ]; then
    echo "ERROR: Run this script from the root of the repository."
    exit 1
fi

BUILD_MODE="normal"
PLATFORM=""

while getopts "p:r" o; do
    case "${o}" in
        p)
            PLATFORM=$OPTARG
            ;;
        r)
            BUILD_MODE="release"
            ;;
        *)
            _usage
            ;;
    esac
done

if [ "$PLATFORM" == "" ]; then
    echo "ERROR: No platform selected."
    exit 1
fi

_build $PLATFORM $BUILD_MODE
