I got this problem from one of our university teachers, and at first I have to admit that it sounded pretty easy. Basically, the statement is as follows:

> Assume that someone is writing on a keyboard, and there are 3 connected clients that listen to the key presses. The rule is that each client has to display _all_ key presses that were received ever since the server started listening, even if the server got closed. One more thing to consider is that each client may not write each key that was pressed instantly, e.g. it's possible that a client takes 1 second to write a single key. Divise an implementation which allows that.

Of course, this statement is a bit general, but it gets the point accross. Right away, it's clear that even if the server closes, the client has to write all remaining keys. 

The catch here is that once the server is closed, it's not possible to read from the TCP connection anymore. And given that the client may take much longer to display those keys, it's simply not a good approach to process each key sequentially, since that would likely cut out a few of the server's key presses.

This gives us the following ideas:

- It's obvious that we have to store the server inputs, and process them separately. Storing a key press happens almost instantly, whereas displaying the key might take longer;
- The order of the keys does matter, so a queue is a good choice to store the keys;
- A worker thread will be processing keys from the queue while the main thread listens for key presses;
- The queue has to be protected against concurrent access by the worker thread and the main thread;
- The client doesn't have to stop the worker thread even if the connection had been closed in the main thread;
- Locking the queue should only be done during writes to the queue, there is no point in locking the queue for the entire processing the worker thread does to display the character. Doing that would effectively be the same as the sequential algorithm.

Obviously, note that the network conditions are not perfect, and it's possible that the network fails, so in this case we assumed perfect network conditions.

Thought: if the client sends the last key and closes the connection almost immediately afterwards, is the last key processed before the connection is perceived as closed?

Running
-------

To run the single process example (multiple processes are simulated through goroutines), just do

```
go run main.go
```

And start typing. Close with an interrupt signal (ctrl + C) or eof condition (ctrl + Z windows, ctrl + D linux). The process will wait for all child threads to complete. To avoid cluttering the terminal, each child thread writes data to a file.

To run the multiple processes example (interaction done with sockets), run the server:

```
go run keyboard.go
```

then wait for the clients to connect. Connect exactly 3 clients by running

```
go run client.go
```

after which the simulation starts again. Proceed just like in previous example in the server terminal. Each client process will write the keys to the terminal, until all are written.

Sample Runs
-----------

Single process, multiple threads:

https://user-images.githubusercontent.com/56697671/225411416-55008bac-9b1b-4b45-8ef3-3dfed8f863ef.mp4

Multiple processes:

https://user-images.githubusercontent.com/56697671/225412094-9294bcb4-de3a-4874-bef0-cb219c6b8fbc.mp4
