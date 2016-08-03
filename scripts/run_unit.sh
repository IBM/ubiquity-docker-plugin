#!/bin/bash
pushd $GOPATH/src/github.ibm.com/almaden-containers/spectrum-container-plugin.git
ginkgo -r -keepGoing -p -trace -randomizeAllSpecs -progress --race \
            -skipPackage=${SKIP_PACKAGES} "$@"
popd
