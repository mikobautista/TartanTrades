import os
import time
import urllib2

NUM_TRADESERVER = 20

os.system("go build tradeservertest.go")
os.system("go build ../server/main/resolverRunner.go")
os.system("go build ../server/main/tradeServerRunner.go")

# Create a resolver
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Create NUM_TRADESERVER trade servers
for i in xrange(NUM_TRADESERVER):
    os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport="+str(1111 + i)+" --tradeport="+str(2222 + i)+" -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
    time.sleep(0.1)

time.sleep(NUM_TRADESERVER/2)

# All trade servers must be registered
expected = ""
for i in xrange(NUM_TRADESERVER):
    expected += "localhost:" + str(1111+i) + ","
os.system("./tradeservertest -hp 'localhost:8888' -n 0 -e " + expected)

# Check that there are no sell requests
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e '' ")

# Create a sell request and check that all trade servers get it
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
for i in xrange(NUM_TRADESERVER):
    os.system("./tradeservertest -hp localhost:" + str(1111+i) + " -n 1 -e '1,1>{}:1;'".format(userid))

# Create two sell requests and check that all trade servers get them
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1112' -x 2 -y 2 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1113' -x 3 -y 3 -token '{}'".format(token))
for i in xrange(NUM_TRADESERVER):
    os.system("./tradeservertest -hp localhost:" + str(1111+i) + " -n 1 -e '1,1>{}:1;2,2>{}:2;3,3>{}:3;'".format(userid, userid, userid))

print "Passed all tests!"
os.system("killall resolverRunner > /dev/null 2>&1")