import os
import urllib
import urllib2
import sys
import csv
import getopt
import random
import getpass
import json
from optparse import OptionParser

def main():
    resolverPort = parseArgs()
    tradeServerPort = getTradeServerPort(resolverPort)
    token = login(resolverPort)
    handleClient(token, resolverPort, tradeServerPort)

def parseArgs():
    try:
        opts, args = getopt.getopt(sys.argv[1:], "resolverPort")
        resolverPort = int(args[0])
        return str(resolverPort)
    except:
        print "Usage: python clientRunner.py <resolverPort>"
        sys.exit(2)

def login(resolverPort):
    while True:
        isLogin = raw_input("Do you already have an account? (Y/N):")
        if isLogin != 'Y' and isLogin != 'N':
            print "Please type 'Y' or 'N'"
            continue
        username = raw_input("Enter your username:")
        password = getpass.getpass("Enter your password:")
        params = {"username" : username, "password" : password}

        if isLogin == 'Y':
            url = "http://localhost:%s/login/?%s" % (resolverPort, urllib.urlencode(params))
            res = urllib2.urlopen(url).read().strip()
            if res == "No such user" or res == "Incorrect Password":
                print "Invalid login credentials. Try again."
                continue
            else:
                return res
        else:
            url = "http://localhost:%s/register/?%s" % (resolverPort, urllib.urlencode(params))
            res = urllib2.urlopen(url).read().strip()
            if res == "User {} Exists".format(username):
                print "Username already exists. Try again."
                continue
            else:
                url = "http://localhost:%s/login?%s" % (resolverPort, urllib.urlencode(params))
                res = urllib2.urlopen(url).read().strip()
                return res

def getTradeServerPort(resolverPort):
    maxTries = 10 # maximum number of tries
    while maxTries > 0:
        url = "http://localhost:%s/servers/" % (resolverPort)
        res = urllib2.urlopen(url).read().strip()
        servers = res.split(',')[:-1]
        tradeServerPort = servers[random.randint(0, len(servers)-1)].split(":")[1]
        url = "http://localhost:%s" % (tradeServerPort)
        try:
            res = urllib2.urlopen(url).read().strip()
            return tradeServerPort
        except:
            maxTries = maxTries - 1
    print "Exceeded maximum number of tries to connect to trade server"
    sys.exit(1)

def handleClient(token, resolverPort, tradeServerPort):
    while True:
        cmd = raw_input("Would you like to buy/sell blocks or view your purchases? If not, press 'q' to quit. (B/S/V/q):")
        if cmd != 'B' and cmd != 'S' and cmd != 'V' and cmd != 'q':
            print "Please type 'B', 'S', or 'q'"
            continue
        if cmd == 'B':
            handleBuy(token, resolverPort, tradeServerPort)
        elif cmd == 'S':
            handleSell(token, resolverPort, tradeServerPort)
        elif cmd == 'V':
            handleView(token, resolverPort, tradeServerPort)
        else:
            print "Thank you for using TartanTrades!"
            sys.exit(0)

def handleBuy(token, resolverPort, tradeServerPort):
    url = "http://localhost:%s/availableblocks/" % (tradeServerPort)
    res = urllib2.urlopen(url).read().strip()
    blocks = res.split(';')[:-1]
    print "\n------------------------------------------------------------------------------------"
    print "\tBlock ID\t|\tLocation\t\t\t|\tSeller"
    print "------------------------------------------------------------------------------------"
    blockIDList = []
    for block in blocks:
        blockIDList.append(getBlockID(block))
        print "\t" + getBlockID(block) + "\t\t|\t" + getBlockLocation(block) + " \t\t|\t" + getSellerName(block, resolverPort)
    print "\n"
    while True:
        blockID = raw_input("Enter the block ID you wish to buy or 'q' to quit:")
        if blockID == "q":
            return
        elif blockID not in blockIDList:
            print "Invalid block ID. Please try again."
            continue
        else:
            params = {"item" : blockID, "token" : token}
            url = "http://localhost:%s/buy/?%s" % (tradeServerPort, urllib.urlencode(params))
            res = urllib2.urlopen(url).read().strip()
            print res
            return

def handleSell(token, resolverPort, tradeServerPort):
    while True:
        location = getLocationWithGeo()
        useGeo = raw_input("It appears you are in {}. Would you like to use this location? (Y/N):".format(location))
        if useGeo != 'Y' and useGeo != 'N':
            print "Please type 'Y' or 'N'"
            continue
        else:
            break

    if useGeo == 'Y':
        (x, y) = getLongLatWithGeo()
    else:
        (x, y) = getLongLatWithoutGeo()
    
    params = {"x" : x, "y" : y, "token" : token}
    url = "http://localhost:%s/sell/?%s" % (tradeServerPort, urllib.urlencode(params))
    res = urllib2.urlopen(url).read().strip()
    print res
    return

def handleView(token, resolverPort, tradeServerPort):
    params = {"token" : token}
    url = "http://localhost:%s/purchases/?%s" % (tradeServerPort, urllib.urlencode(params))
    res = urllib2.urlopen(url).read().strip()
    blocks = res.split(';')[:-1]
    print "\n------------------------------------------------------------------------------------"
    print "\tBlock ID\t|\tLocation\t\t\t|\tSeller"
    print "------------------------------------------------------------------------------------"
    blockIDList = []
    for block in blocks:
        blockIDList.append(getBlockID(block))
        print "\t" + getBlockID(block) + "\t\t|\t" + getBlockLocation(block) + "\t\t|\t" + getSellerName(block, resolverPort)
    print "\n"

def getLocationWithGeo():
    res = urllib2.urlopen("http://freegeoip.net/json").read().strip()
    obj = json.loads(res)
    return obj["city"] + ", " + obj["region_code"] + " " + obj["zipcode"]

def getLongLatWithGeo():
    res = urllib2.urlopen("http://freegeoip.net/json").read().strip()
    obj = json.loads(res)
    return (obj["longitude"], obj["latitude"])

def getLongLatWithoutGeo():
    while True:
        x = raw_input("Enter longitude of selling point:")
        if isNumber(x):
            if float(x) <= -180 or 180 <= float(x):
                print "Longitude must be in range [-180, 180]"
            else:
                break
        else:
            print "Longitude must be a number."
    while True:
        y = raw_input("Enter latitude of selling point:")
        if isNumber(y):
            if -90 <= float(y) and float(y) <= 90:
                break
            else:
                print "Latitude must be in range [-90, 90]"
        else:
            print "Latitude must be a number."
    return (format(float(x), '.4f'), format(float(y), '.4f'))

def getBlockID(block):
    return block.split(":")[1]

def getSellerName(block, resolverPort):
    userID = block.split(">")[1].split(":")[0]
    params = {"id" : userID}
    url = "http://localhost:%s/lookup/?%s" % (resolverPort, urllib.urlencode(params))
    res = urllib2.urlopen(url).read().strip()
    return res

def getBlockLocation(block):
    return "(" + block.split(">")[0] + ")"

def isNumber(s):
    try:
        float(s) # for int, long and float
    except ValueError:
        try:
            complex(s) # for complex
        except ValueError:
            return False
    return True

main()