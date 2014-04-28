import os
import time
import urllib2

token = "eyJVc2VybmFtZSI6Im1hdCIsIkV4cGVyYXRpb24iOiIyMDE0LTA0LTI4VDAwOjM4OjQ5LjQyMDA2OTkzOS0wNDowMCJ9"

os.system("go build tradeservertest.go")
os.system("go build ../server/main/resolverRunner.go")
os.system("go build ../server/main/tradeServerRunner.go")

# Create a resolver and a tradeserver runner
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false > /dev/null 2>&1)&")
time.sleep(1)
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Tradeserver must be registered with resolver
os.system("./tradeservertest -hp 'localhost:8888' -n 0 -e 'localhost:1111,'")

# No sell requests at the beginning
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e '' ")

# Create a sell request and check that it gets registered
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e '1,1>{}:1;'".format(userid))

print "Passed all tests!"
os.system("killall resolverRunner > /dev/null 2>&1")