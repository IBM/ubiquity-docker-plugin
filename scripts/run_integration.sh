#!/bin/bash
pushd $SPECTRUM_DIR/src/gitlabhost.rtp.raleigh.ibm.com/spectrum-plugin/main/
ginkgo -r -keepGoing -p -trace -randomizeAllSpecs -progress --race "$@"
popd
