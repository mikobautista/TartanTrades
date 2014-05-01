import os
import time
import urllib2
import sys
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-v", action="store_true", dest="verbose", default=False)
args = parser.parse_args()
VERBOSE = args[0].verbose
db_user = args[1][0]
db_pw = args[1][1]

NUM_TRADESERVER = 4 # must be at least 3

os.system("go build ../go/main/tradeservertest.go")
os.system("go build ../../server/main/resolverRunner.go")
os.system("go build ../../server/main/tradeServerRunner.go")

# Create a resolver
if VERBOSE: print "Starting 1 resolver and 4 tradeservers..."
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user={} -db_pw={} > /dev/null 2>&1)&".format(db_user, db_pw))
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Create NUM_TRADESERVER trade servers
for i in xrange(NUM_TRADESERVER):
    os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport="+str(1111 + i)+" --tradeport="+str(2222 + i)+" -dropTableOnStart=true -createTableOnStart=true -db_user={} -db_pw={} > /dev/null 2>&1)&".format(db_user, db_pw))
    time.sleep(0.1)

time.sleep(NUM_TRADESERVER/2)

# All trade servers must be registered
if VERBOSE: print "Checking that all tradeservers get registered..."
expected = ""
for i in xrange(NUM_TRADESERVER):
    expected += "localhost:" + str(1111+i) + ","
os.system("./tradeservertest -hp 'localhost:8888' -n 0 -e " + expected)

# Check that there are no sell requests
if VERBOSE: print "Checking that there are no sell requests initially..."
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e '' ")

# Create a sell request and check that all trade servers get it
if VERBOSE: print "Creating a sell request..."
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
if VERBOSE: print "Checking that sell request gets registered..."
for i in xrange(NUM_TRADESERVER):
    os.system("./tradeservertest -hp localhost:" + str(1111+i) + " -n 1 -e '1,1>{}:1;'".format(userid))

# Create two sell requests and check that all trade servers get them
if VERBOSE: print "Creating 2 sell requests one after the other..."
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1112' -x 2 -y 2 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1112' -x 3 -y 3 -token '{}'".format(token))
time.sleep(1) # wait to be registered
if VERBOSE: print "Checking that both sell requests get registered in the same order by all tradeservers..."
for i in xrange(NUM_TRADESERVER):
    os.system("./tradeservertest -hp localhost:" + str(1111+i) + " -n 1 -e '1,1>{}:1;2,2>{}:2;3,3>{}:3;'".format(userid, userid, userid))

os.system("killall resolverRunner > /dev/null 2>&1")
os.system("killall tradeServerRunner > /dev/null 2>&1")