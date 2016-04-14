#!/bin/bash
cd $SPECTRUM_DIR
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega
go get github.com/maxbrunsfeld/counterfeiter
git submodule update --init --recursive
cd ..
