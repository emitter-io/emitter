# Emitter Stats
This golang package is used for monitoring [emitter.io](emitter.io) cluster. This provides a tight binary compression and a single histogram/counter abstraction in order to deal with various kinds of system monitoring. This package is compatible with GopherJS (snapshot-only) and hence can also be compiled to javascript. Documentation is available on [go doc](https://godoc.org/github.com/emitter-io/stats).

[![Join the chat at https://gitter.im/emitter-io/public](https://badges.gitter.im/emitter-io/public.svg)](https://gitter.im/emitter-io/public?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) 
[![Build status](https://ci.appveyor.com/api/projects/status/3y2d9ssq760g8bfd?svg=true)](https://ci.appveyor.com/project/Kelindar/stats)
[![Coverage Status](https://coveralls.io/repos/github/emitter-io/stats/badge.svg?branch=master)](https://coveralls.io/github/emitter-io/emitter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/emitter-io/stats)](https://goreportcard.com/report/github.com/emitter-io/stats)

## Installation

```
go get -u github.com/emitter-io/stats
```

## Quick Start

The package itself provides a general-purpose monitoring capabilities, with tight encoding using our [binary codec](https://github.com/kelindar/binary). While it's primarily have been built for emitter, it can be used anywhere.

Typical usage consists of creating a metric container, measuring various metrics and sending snapshots over the wire.

```
rand.Seed(time.Now().UnixNano())

// Create a container
m := stats.New()

// Measure few metrics
m.Measure("my.metric.1", rand.Int31n(1000))
m.Measure("my.metric.2", rand.Int31n(1000))

// Create a snapshot which can be transferred over the wire
bytes := m.Snapshot()

// Restore a snapshot from binary
v, err := stats.Restore(bytes)

// Get the values back
percentiles := v[0].Quantile(50, 90, 95, 99)
average := v[0].Mean()
count := v[0].Count()
```
