#!/bin/bash
SKIP_PACKAGES=main
pushd $SPECTRUM_DIR/src/gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin
ginkgo -r -keepGoing -p -trace -randomizeAllSpecs -progress --race \
            -skipPackage=${SKIP_PACKAGES} "$@"
popd
