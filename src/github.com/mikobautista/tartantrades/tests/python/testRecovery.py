import os
import time
import urllib2
import sys
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-v", action="store_true", dest="verbose", default=False)
args = parser.parse_args()
VERBOSE = args[0].verbose

# Create a resolver
if VERBOSE: print "Creating 1 resolver and 2 tradeservers..."
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Start two trade servers
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true -db_user=root -db_pw=password > /dev/null 2>&1)&")
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(1)

# Create two sell requests
if VERBOSE: print "Creating 2 sell requests one after the other..."
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1111' -x 2 -y 2 -token '{}'".format(token))
time.sleep(2) # wait for changes to be committed

# Kill the second trade server
if VERBOSE: print "Killing the 2nd trade server"
os.system("./tradeservertest -hp localhost:1112 -n 3 -e 'Stopping trade server in 5 seconds'")
time.sleep(5) # wait for trade server to die

# Create one more sell request and a buy request
if VERBOSE: print "Creating 1 sell request and 1 buy request..."
os.system("./tradeservertest -n 2 -e 'FAIL: Paxos majority timeout' -hp 'localhost:1111' -x 3 -y 3 -token '{}'".format(token))
os.system("./tradeservertest -n 4 -e 'FAIL: Paxos majority timeout' -hp 'localhost:1111' -item 2 -token '{}'".format(token))
time.sleep(2)

# Revive the trade server and check that it contains all requests
if VERBOSE: print "Reviving the tradeserver..."
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(3) # wait for trade server to recover
if VERBOSE: print "Checking that tradeserver properly recovered..."
os.system("./tradeservertest -hp localhost:1112 -n 1 -e '1,1>{}:1;2,2>{}:2;'".format(userid, userid))

os.system("killall resolverRunner > /dev/null 2>&1")
os.system("killall tradeServerRunner > /dev/null 2>&1")