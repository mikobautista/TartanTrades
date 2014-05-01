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
if VERBOSE: print "Starting 1 resolver..."
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Start the early trade server
if VERBOSE: print "Starting 1 early tradeserver..."
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(1)

# Create three sell requests
if VERBOSE: print "Creating 3 sell requests..."
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1111' -x 2 -y 2 -token '{}'".format(token))
os.system("./tradeservertest -n 2 -e 'OK' -hp 'localhost:1111' -x 3 -y 3 -token '{}'".format(token))
time.sleep(2) # wait for changes to be committed

# Start the late trade server
if VERBOSE: print "Starting 1 late tradeserver..."
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1112 --tradeport=2223 -dropTableOnStart=true -createTableOnStart=true -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(2) # wait for trade server to recover

# Check that the late trade server also registered the request
if VERBOSE: print "Checking that the late trade server registered all 3 sell requests..."
os.system("./tradeservertest -hp localhost:1112 -n 1 -e '1,1>{}:1;2,2>{}:2;3,3>{}:3;'".format(userid, userid, userid))

os.system("killall resolverRunner > /dev/null 2>&1")
os.system("killall tradeServerRunner > /dev/null 2>&1")