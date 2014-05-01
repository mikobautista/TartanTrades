#!/bin/bash
go run ./resolverRunner.go -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user=$1 -db_pw=$2
