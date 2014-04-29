import os
import time
import urllib2
import sys
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-v", action="store_true", dest="verbose", default=False)
args = parser.parse_args()
VERBOSE = args[0].verbose

token = "eyJVc2VybmFtZSI6Im1hdCIsIkV4cGVyYXRpb24iOiIyMDE0LTA0LTI4VDAwOjM4OjQ5LjQyMDA2OTkzOS0wNDowMCJ9"

# Create a resolver and a tradeserver runner
if VERBOSE: print "Starting 1 resolver and 1 tradeserver..."
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false > /dev/null 2>&1)&")
time.sleep(1)
os.system("(./tradeServerRunner -resolverHost=127.0.0.1 -resolverHttpPort=8888 -resolverTcpPort=1234 --httpport=1111 --tradeport=2222 -dropTableOnStart=true -createTableOnStart=true > /dev/null 2>&1)&")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=foo&password=bar").read().strip()
userid = urllib2.urlopen("http://localhost:8888/validate/?token={}".format(token)).read().strip()

# Tradeserver must be registered with resolver
if VERBOSE: print "Checking that tradeserver is registered..."
os.system("./tradeservertest -hp 'localhost:8888' -n 0 -e 'localhost:1111,'")

# No sell requests at the beginning
if VERBOSE: print "Checking that there are no sell requests initially..."
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e '' ")

# Create a sell request and check that it gets registered
if VERBOSE: print "Creating a sell request..."
os.system("./tradeservertest -n 2 -e '' -hp 'localhost:1111' -x 1 -y 1 -token '{}'".format(token))
if VERBOSE: print "Checking that sell request is registered..."
os.system("./tradeservertest -hp 'localhost:1111' -n 1 -e '1,1>{}:1;'".format(userid))

os.system("killall resolverRunner > /dev/null 2>&1")