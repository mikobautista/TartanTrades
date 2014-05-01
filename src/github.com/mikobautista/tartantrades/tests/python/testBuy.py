import os
import time
import urllib2
import sys
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-v", action="store_true", dest="verbose", default=False)
args = parser.parse_args()
VERBOSE = args[0].verbose

# Create a resolver and two tradeserver runners
if VERBOSE: print "Starting 1 resolver and 2 tradeservers..."
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(1)
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true -db_user=root -db_pw=password > /dev/null 2>&1)&")
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Create a sell request
if VERBOSE: print "Creating 1 sell request on the second tradeserver..."
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
time.sleep(1)

# Create a buy request for the second item on the other server
if VERBOSE: print "Creating a buy requeston the first tradeserver..."
os.system("./tradeservertest -n 4 -e 'OK' -hp 'localhost:1111' -item 1 -token '{}'".format(token))
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e ''".format(userid))
time.sleep(1)

# Try to buy an invalid item as well
if VERBOSE: print "Attempting to buy an invalid item..."
os.system("./tradeservertest -n 4 -e 'Invalid Item' -hp 'localhost:1111' -item 10 -token '{}'".format(token))

os.system("killall resolverRunner > /dev/null 2>&1")
os.system("killall tradeServerRunner > /dev/null 2>&1")