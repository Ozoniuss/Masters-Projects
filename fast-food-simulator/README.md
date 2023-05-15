# Fast food simulator

This app is a simple fast food simulator for one of my master's project. It essentially allows you to:

- Order something to eat, which gives you an order number
- Get your order, with your ticket number
- Get real-time updates on the order

Overall design
--------------

The application exposes a simple http server, with an endpoint for placing an order, retrieving the order when it is finished and viewing the order updates. This is based on the ticket number system employed at various fast-foods: you make an order, which gives you a ticket, then you wait for your order to be finished. In the meantime, you can see whether the order is started, preparing or finished on a board.

The `/order` endpoint gives you a ticket number (say `3`), and places the order in an order queue. The initial status is `TAKEN`. Once previous orders have been processed and a fast food worker is available to start preparing the order, the status goes to `PREPARING`. When the order is finished, the status is changed to `READY` and the `/take?order=3` can be called to retrieve the order.

Orders are stored in a (toy) database, which enables access to the orders accross multiple processes. This was not required for running the workers in different threads, but for compatibility and not wanting to write more code, the approach allows for both running the workers in separate threads or processes. See below the difference between each running method.

Ticket numbers are generated for simplicity in increasing order. See [Atomic Counter](#atomic-counter) for more details on how that is implemented.

Real-time updates can be used in order to make a visual representation of the order status, similar to the panels from real fast food restaurants. To get real-time notifications, call the endpoint `/updates` which will send all orders with their statuses whenever something changes through _Server-Sent Events_.

RabbitMQ is used as the message broker which holds the orders. This is a well-established reliable message broker which enables message persistence during server failures, and delivery guarantees. If a worker fails during the processing of an order, the order is requeued (TODO: with a high priority).

There are two ways to run this thing, either the simple way if you're lazy, or the complex way if you want to have separate processes for taking orders and preparing orders, which simulates more of a real-world scenario. The implementation differences are described below, see [Simple Example](#simple-example) and [Complex Example](#complex-example). The main difference between the two is how order notifications are send, in the simple example through a channel and in the complex example through a queue. Note that orders are still placed in a queue in both cases, but in the simple example a channel would have sufficed for that (at the cost of reduced failure safety).

Storing Orders
--------------

The database is simulataed by a file, in json format. _This is obviously extremely inneficient and probably no one would ever do that even for a personal project_, but as a proof of concept is fine since the assignment focused on interacting with message queues and I didn't want to add additional complexity. Every write or update operations reads the entire content of the file and then overwrites the new content to the file, while the file provides a read-level lock and a write-level lock for the entire file (kindof like how mongodb did it at first). There is no point in trying to optimize this, the approach is flawed by design and databases have already spent decades optimizing the process. But, for a demo it's quick to use, no additional dependencies, and the json format makes it a lot easier to debug.

Atomic Counter
--------------

The atomic counter I did store in binary format, simply because it was easier to know that I always had to store exactly 4 bytes, and with a binary viewer extension you can see exactly what number you have in there. I stored it in Big Endian format.

Simple Example
--------------

Complex Example
---------------

Running
-------