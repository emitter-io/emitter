# Emitter address
This golang package is used for IP/hardware address parsing for [emitter.io](emitter.io). This provides a simple way of parsing a variety of IP addresses and, in future, domain names.

[![Join the chat at https://gitter.im/emitter-io/public](https://badges.gitter.im/emitter-io/public.svg)](https://gitter.im/emitter-io/public?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) 
[![Build status](https://ci.appveyor.com/api/projects/status/3y2d9ssq760g8bfd?svg=true)](https://ci.appveyor.com/project/Kelindar/address)
[![Coverage Status](https://coveralls.io/repos/github/emitter-io/address/badge.svg?branch=master)](https://coveralls.io/github/emitter-io/address?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/emitter-io/address)](https://goreportcard.com/report/github.com/emitter-io/address)

## Installation

```
go get -u github.com/emitter-io/address
```

## API Documentation

For full API documentation, please refer to our [godoc.org/github.com/emitter-io/address](https://godoc.org/github.com/emitter-io/address) which contains all of the methods that package exposes.

## Parsing IP Addresses

Typical usage consists of calling parse function which provides a set of helpers such as (e.g. `private`, `public` and `external`) which then get appropriately parsed to the TCP IP Address, as demonstrated below.

```
addr, err := address.Parse("external:8080")
```
