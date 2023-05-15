# Fast food simulator

This app is a simple fast food simulator for one of my master's project. It essentially allows you to:

- Order something to eat, which gives you an order number
- View all the orders, based on status: order taken, preparing, order finished
- Get your order, with your ticket number


Storing orders and the atomic counter
-------------------------------------

The database is simulataed by a file, in json format. _This is obviously extremely inneficient and probably no one would ever do that even for a personal project_, but as a proof of concept is fine since the assignment focused on interacting with message queues and I didn't want to add additional complexity. Every write or update operations reads the entire content of the file and then overwrites the new content to the file, while the file provides a read-level lock and a write-level lock for the entire file (kindof like how mongodb did it at first). There is no point in trying to optimize this, the approach is flawed by design and databases have already spent decades optimizing the process. But, for a demo it's quick to use, no additional dependencies, and the json format makes it a lot easier to debug.

The atomic counter I did store in binary format, simply because it was easier to know that I always had to store exactly 4 bytes, and with a binary viewer extension you can see exactly what number you have in there. I stored it in Big Endian format.