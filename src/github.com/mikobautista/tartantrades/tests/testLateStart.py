import os
import time
import urllib2

os.system("go build tradeservertest.go")
os.system("go build ../server/main/resolverRunner.go")
os.system("go build ../server/main/tradeServerRunner.go")

# Create a resolver
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Start the early trade server
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
time.sleep(1)

# Create three sell requests
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 2 -y 2 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 3 -y 3 -token '{}'".format(token))
time.sleep(2) # wait for changes to be committed

# Start the late trade server
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
time.sleep(2) # wait for trade server to recover

# Check that the late trade server also registered the request
os.system("./tradeservertest -hp localhost:1112 -n 1 -e '1,1>{}:1;2,2>{}:2;3,3>{}:3;'".format(userid, userid, userid))

print "Passed all tests!"
os.system("killall resolverRunner > /dev/null 2>&1")