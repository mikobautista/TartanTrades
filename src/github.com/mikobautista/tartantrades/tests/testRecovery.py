import os
import time
import urllib2

token = "eyJVc2VybmFtZSI6Im1hdCIsIkV4cGVyYXRpb24iOiIyMDE0LTA0LTI4VDAwOjM4OjQ5LjQyMDA2OTkzOS0wNDowMCJ9"

os.system("go build tradeservertest.go")
os.system("go build ../server/main/resolverRunner.go")
os.system("go build ../server/main/tradeServerRunner.go")

# Create a resolver
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Start two trade servers
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
time.sleep(1)

# Create two sell requests
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 2 -y 2 -token '{}'".format(token))
time.sleep(2) # wait for changes to be committed

# Kill the second trade server
os.system("./tradeservertest -hp localhost:1112 -n 3 -e 'Stopping trade server in 5 seconds'")
time.sleep(5) # wait for trade server to die

# Create one more sell request and a buy request
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 3 -y 3 -token '{}'".format(token))
os.system("./tradeservertest -n 4 -e '' -hp 'localhost:1111' -item 2 -token '{}'".format(token))
time.sleep(2)

# Revive the trade server and check that it contains all requests
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
time.sleep(3) # wait for trade server to recover
os.system("./tradeservertest -hp localhost:1112 -n 1 -e '1,1>{}:1;3,3>{}:3;'".format(userid, userid))

print "Passed all tests!"
os.system("killall resolverRunner > /dev/null 2>&1")