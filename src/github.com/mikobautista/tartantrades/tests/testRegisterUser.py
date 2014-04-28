import os
import time
import urllib2

os.system("go build tradeservertest.go")
os.system("go build ../server/main/resolverRunner.go")
os.system("go build ../server/main/tradeServerRunner.go")

# Create a resolver 
os.system("(./resolverRunner -tradeport=1234 -httpport=8888 -checkSessionExperation=false > /dev/null 2>&1)&")
time.sleep(1)

# Create a new account
os.system("./tradeservertest -n 5 -e 'User miko Created!' -hp 'localhost:8888' -user 'miko' -pass 'pw'")
time.sleep(1)
os.system("./tradeservertest -n 5 -e 'User miko Exists' -hp 'localhost:8888' -user 'miko' -pass 'pw'")
time.sleep(1)

token = urllib2.urlopen("http://localhost:8888/login/?username=miko&password=pw").read().strip()

# Delete the account
os.system("./tradeservertest -n 6 -e 'User Deleted' -hp 'localhost:8888' -token {}".format(token))
time.sleep(1)
os.system("./tradeservertest -n 6 -e 'Invalid token' -hp 'localhost:8888' -token {}".format(token))
time.sleep(1)

print "Passed all tests!"
os.system("killall resolverRunner > /dev/null 2>&1")