#!/bin/bash
go run ./tradeServerRunner.go -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=$1 --tradeport=$2 -dropTableOnStart=true -createTableOnStart=true -db_user=$3 -db_pw=$4
