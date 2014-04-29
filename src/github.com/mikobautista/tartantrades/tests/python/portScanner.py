import socket
from optparse import OptionParser

parser = OptionParser()
parser.add_option("-v", action="store_true", dest="verbose", default=False)
args = parser.parse_args()
VERBOSE = args[0].verbose

NUM_TRADESERVER = 9 # must be at least 3

portList = []

# Add in httpports and tcpports for resolver and tradeservers
portList.append(8888)
portList.append(1234)
for i in xrange(NUM_TRADESERVER):
    portList.append(1111+i)
    portList.append(2222+i)

# Ensure that all the ports are closed
for port in portList:
    # print port
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    result = s.connect_ex(('127.0.0.1', port))
    if result == 0:
        if VERBOSE: print "Closing socket on port " + str(port)
    s.close()