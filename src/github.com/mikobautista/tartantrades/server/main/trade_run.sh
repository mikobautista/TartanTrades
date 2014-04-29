#!/bin/bash
go run ./tradeServerRunner.go -resolverHost=127.0.0.1 -resolverHttpPort=1234 -resolverTcpPort=1234 --httpport=$1 --tradeport=$2 -dropTableOnStart=true -createTableOnStart=true
