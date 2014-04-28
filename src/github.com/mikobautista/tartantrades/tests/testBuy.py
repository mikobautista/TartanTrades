import os
import time
import urllib2

token = "eyJVc2VybmFtZSI6ImZvbyIsIkV4cGVyYXRpb24iOiIyMDE0LTA0LTI4VDA1OjIwOjAwLjc0MTg0NjgyOC0wNDowMCJ9"

os.system("go build tradeservertest.go")
os.system("go build ../server/main/resolverRunner.go")
os.system("go build ../server/main/tradeServerRunner.go")

# Create a resolver and two tradeserver runners
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false > /dev/null 2>&1)&")
time.sleep(1)
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Create two sell requests
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1112' -x 2 -y 2 -token '{}'".format(token))
time.sleep(1)

# Create a buy request for the second item on the other server
os.system("./tradeservertest -n 4 -e '' -hp 'localhost:1111' -item 2 -token '{}'".format(token))
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e '1,1>{}:1;'".format(userid))
time.sleep(1)

# Try to buy an invalid item as well
os.system("./tradeservertest -n 4 -e 'Invalid Item' -hp 'localhost:1111' -item 10 -token '{}'".format(token))

print "Passed all tests!"
os.system("killall resolverRunner > /dev/null 2>&1")