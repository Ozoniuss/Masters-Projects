import socket
import struct
import random

SERVER_HOST = "localhost"
SERVER_PORT = 9987
SERVER_TYPE = socket.SOCK_STREAM

CLIENT_HELLO = 0
SERVER_HELLO = 1
CLIENT_DONE = 2
SERVER_DONE = 3
ERROR = 4
CLIENT_MSG = 5
SERVER_MSG = 6
CLIENT_CLOSE = 7
SERVER_CLOSE = 8

class ConnState:
    def __init__(self):
        self.pubKey = None
        self.symKey = None

def newState():
    return ConnState()

def readMessage():
    typ, msg = map(str, input().split(':'))
    return int(typ), msg

def writeMsg(typ, msg, s):
    sends = bytearray([typ])
    if s.symKey is not None and msg != "":
        msg = crypt.EncryptAES(s.symKey[:], bytearray(msg, 'utf-8'))
    sends += bytearray(msg, 'utf-8')
    return sends

def processMessage(connection, s):
    buffer = connection.recv(1024)
    if len(buffer) == 0:
        return Exception("Received null message")
    header = buffer[0]
    content = buffer[1:]

    if header == SERVER_HELLO:
        print("[server hello] received server hello")

        pubKey = &crypt.PublicKey{}
        pubKey.Unmarshal(content)

        s.pubKey = pubKey

        print("[server hello] public key is ", pubKey)

        symKey = generateSymKey()
        print("[server hello] generated sym key: ", symKey)

        msg = pubKey.EncryptString(symKey[:])

        connection.send(writeMsg(CLIENT_DONE, msg, s))

        s.symKey = &symKey

    elif header == SERVER_MSG:
        print("[message] server encrypted message as: ", content)

    elif header == SERVER_DONE:
        print("[server done] handshake complete")

    elif header == ERROR:
        print("[error] received error: ", content)

    else:
        print("[error] handshake complete")

def generateSymKey():
    key = bytearray(random.getrandbits(8) for i in range(32))
    return key

def main():
    #establish connection
    connection = socket.socket(socket.AF_INET, SERVER_TYPE)
    connection.connect((SERVER_HOST, SERVER_PORT))

    state = newState()

    while True:
        typ, msg = readMessage()
        sends = writeMsg(typ, msg, state)
        connection.send(sends)
        processMessage(connection, state)

if __name__ == '__main__':
    main()