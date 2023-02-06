# Safeecho

This is a cool project I've done with one of my friends as the final assignment for our cryptography class. The idea behind the project is to attempt simulating the TLS protocol in a simplified way: a server which defines a various number of possible accepted messages that allows updating the state of the connection and performing the TLS handshake, and a client that automatically performs the TLS handshake and allows writing messages. Similar to an "echo" server, the server returns the messages it received to the client, except that it returns them exactly how they were encoded when it received them.

The catch of this assignment was that we had only two days to complete it, alongside full-time work. In order to finish it on time, we simplified the TLS protocol to:

- The client sends a client hello message;
- The server generates a public RSA public/private key pair, and sends the public key to the client;
- The client generates a private symmetric key, encrypts it with the public key and sends it back to the server;
- The server decrypts the symmetric key, and sends a server done message;
- Communication begins, being encrypted with the symmetric key.

For the symmetric key encryption, we've used Golang's AES library. The RSA algorithm, has been implemented completely by hand, including the public-private key generation and operations on big integers, much higher than 64 bits. Since we like low-level programming, we went with opening a tcp socket and storing bare tcp connections. Of course, TLS is most commonly used to encrypt communication via the HTTP protocol, but HTTP communication was not relevant for the purpose of this application.

Coding this was nice and we're evolving it to become our own chatting application, in order to improve the security of our online communication. See the continuation of the project at this link: https://github.com/theshamans/safechat