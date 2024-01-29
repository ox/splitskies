#!/usr/bin/env bash

GITROOT=$(git rev-parse --show-toplevel)

set -e

pushd $GITROOT
reflex -c reflex.conf
popd