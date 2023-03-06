This is an implementation of the job handler algorithm from the book "Introduction to Reliable and Secure Distributed Programming", both the synchronous version and the asynchronous version. 

Synchronous
-----------

Run the synchronous version (algorithm 1.1) using 

```
go run synchronous.go
```

Note that the jobs are confirmed in the same order that they have been received. There are two caveats:

1. This is a simulation on a single process, so it can only do so much. The ability to print the jobs in the order that they got processed is limited by the ability of the print calls to be serialized correctly in the goroutines. This always happens since it takes more than 1s to complete a job and they're processed synchronously, but note that if two jobs were completed at (approximately) the same time, even if job 1 completes before job 2, _printing_ that job 1 completed might happen before printing that job 2 completed.

2. Note that the confirmation is not always printed once the job is completed. This is also because things are happening concurrently: the processor might schedule the printing of working on the next job before printing that the previous job completed in the other thread. In a distributed system, this would not be the case because the system would only print the submissions it receives locally, and each node sending jobs would get their confirmation indepentend of the other nodes.

Asynchronous
------------

Run the asynchronous version (algorithm 1.2) using 

```
go run asynchronous.go
```

Note how quickly the jobs are confirmed, but it takes a while for them to execute. Obviously, this doesn't bring any improvements over the synchronous implementation, jobs are still being processed synchronously by the job handler, except that the confirmation that the job will eventually be processed is sent before the actual processing. In practice, jobs could be placed in a processing queue, and be processed by individual workers reading from that queue.