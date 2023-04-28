# go-apache-log-parser

## Description

go-apache-log-parser is a parser for Apache log files written in Go.

This has been forked from [ksaylor11/go-apache-log-parser](https://github.com/ksaylor11/go-apache-log-parser) who forked it from [dsparling/go-apache-log-parser](https://github.com/dsparling/go-apache-log-parser) and modified with the changes below.

## Changes

* Updated to use Go modules
* Updated test to use included test log file
* Added the ability to parse gzipped tar (tar.gz / .tgz) files
* Added the ability to parse gzipped (.gz) files

## Usage

Importing:

    import (
        apachelogparser "github.com/timbutler/go-apache-log-parser"
    )

Example:

    lines, err := apachelogparser.Parse("logs/access_log")
    if err != nil {
        log.Fatal(err)
    }
    for _, line := range lines {
        fmt.Printf("remote host: %s\n", line.RemoteHost)
        fmt.Printf("time: %s\n", line.Time)
        fmt.Printf("request: %s\n", line.Request)
        fmt.Printf("status: %d\n", line.Status)
        fmt.Printf("bytes: %d\n", line.Bytes)
        fmt.Printf("referer: %s\n", line.Referer)
        fmt.Printf("user agent: %s\n", line.UserAgent)
        fmt.Printf("url: %s\n", line.URL)
        fmt.Printf("method: %s\n", line.Method)
        fmt.Printf("protocol: %s\n", line.Protocol)
    }
