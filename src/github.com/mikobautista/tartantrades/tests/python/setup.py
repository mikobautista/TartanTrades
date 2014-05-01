import os
import time
import urllib2
import sys
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-v", action="store_true", dest="verbose", default=False)
args = parser.parse_args()
VERBOSE = args[0].verbose

os.system("go run ../../scripts/main/create_resolver_tables.go")
time.sleep(2)
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false -db_user=root -db_pw=password > /dev/null 2>&1)&")
time.sleep(1)
urllib2.urlopen("http://localhost:8888/register/?username=foo&password=bar")
time.sleep(1)

os.system("killall resolverRunner > /dev/null 2>&1")