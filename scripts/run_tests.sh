#!/bin/bash

pushd $HOME/go/src/code.cloudfoundry.org/credhub-cli
  unset CREDHUB_DEBUG
  make test
popd
