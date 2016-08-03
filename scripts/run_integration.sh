#!/bin/bash
pushd $GOPATH/src/github.ibm.com/almaden-containers/spectrum-container-plugin.git
ginkgo -r -keepGoing -trace -randomizeAllSpecs -progress --race "$@"
popd
