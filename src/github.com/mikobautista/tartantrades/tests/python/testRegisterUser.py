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

# Create a resolver 
if VERBOSE: print "Starting 1 resolver..."
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user={} -db_pw={} > /dev/null 2>&1)&".format(db_user, db_pw))
time.sleep(1)

# Create a new account
if VERBOSE: print "Creating 1 new account..."
os.system("./tradeservertest -n 5 -e 'User miko Created!' -hp 'localhost:8888' -user 'miko' -pass 'pw'")
time.sleep(1)

if VERBOSE: print "Attempting to create a new account with the same username..."
os.system("./tradeservertest -n 5 -e 'User miko Exists' -hp 'localhost:8888' -user 'miko' -pass 'pw'")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=miko&password=pw").read().strip()

# Delete the account
if VERBOSE: print "Delete the account..."
os.system("./tradeservertest -n 6 -e 'User Deleted' -hp 'localhost:8888' -token {}".format(token))
time.sleep(1)

if VERBOSE: print "Attempting to delete the same account again..."
os.system("./tradeservertest -n 6 -e 'Invalid token' -hp 'localhost:8888' -token {}".format(token))
time.sleep(1)

os.system("killall resolverRunner > /dev/null 2>&1")
os.system("killall tradeServerRunner > /dev/null 2>&1")