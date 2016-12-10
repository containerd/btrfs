# go-btrfs
[![GoDoc](https://godoc.org/github.com/stevvooe/go-btrfs?status.svg)](https://godoc.org/github.com/stevvooe/go-btrfs) [![Build Status](https://travis-ci.org/stevvooe/go-btrfs.svg?branch=master)](https://travis-ci.org/stevvooe/go-btrfs)

Native Go bindings for btrfs.

As opposed to other drivers, this driver pre-generates native Go definitions,
rather than using cgo.

# Status

These are in the early stages. We will try to maintain stability, but please
vendor if you are relying on these directly.

# Contribute

This package may not cover all the use cases for btrfs. If something you need
is missing, please don't hesitate to submit a PR.
