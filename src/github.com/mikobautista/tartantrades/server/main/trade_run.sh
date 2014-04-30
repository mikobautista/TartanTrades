#!/bin/bash
go run ./tradeServerRunner.go -resolverHost=$1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=$2 --tradeport=$3 -dropTableOnStart=true -createTableOnStart=true
