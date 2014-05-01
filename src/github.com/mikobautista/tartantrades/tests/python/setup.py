import os
import time
import urllib2
import sys
import socket
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-v", action="store_true", dest="verbose", default=False)
args = parser.parse_args()
VERBOSE = args[0].verbose
db_user = args[1][0]
db_pw = args[1][1]

NUM_TRADESERVER = 4 # must be at least 3

portList = []

# Add in httpports and tcpports for resolver and tradeservers
portList.append(8888)
portList.append(1234)
for i in xrange(NUM_TRADESERVER):
    portList.append(1111+i)
    portList.append(2222+i)

# Ensure that all the ports are closed
for port in portList:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    result = s.connect_ex(('127.0.0.1', port))
    if result == 0:
        if VERBOSE: print "Closing socket on port " + str(port)
    s.close()

time.sleep(1)

os.system("go run ../../scripts/main/create_resolver_tables.go")
time.sleep(2)
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user={} -db_pw={} > /dev/null 2>&1)&".format(db_user, db_pw))
time.sleep(1)
urllib2.urlopen("http://localhost:8888/register/?username=foo&password=bar")
time.sleep(1)

os.system("killall resolverRunner > /dev/null 2>&1")