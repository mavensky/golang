# Overview
We are given total possible page numbers that can be referred.
We are also given cache (or memory) size (Number of page frames that cache can hold at a time).
The LRU caching scheme is to remove the least recently used frame when the cache is full and a new page is referenced which is not there in cache.
Please see the Galvin book for more details (see the LRU page replacement slide here).

We use two data structures to implement an LRU Cache.
1. Queue
   which is implemented using a doubly linked list. The maximum size of the queue will be equal to the total number of frames available (cache size).
   The most recently used pages will be near front end and least recently pages will be near rear end.
2. Hash
   A Hash with page number as key and address of the corresponding queue node as value.

When a page is referenced, the required page may be in the memory.
If it is in the memory, we need to detach the node of the list and bring it to the front of the queue.
If the required page is not in the memory, we bring that in memory.
In simple words, we add a new node to the front of the queue and update the corresponding node address in the hash.
If the queue is full, i.e. all the frames are full, we remove a node from the rear of queue, and add the new node to the front of queue.

# Example
Consider the following reference string
`1, 2, 3, 4, 1, 2, 5, 1, 2, 3, 4, 5`

Find the number of page faults using least recently used (LRU) page replacement algorithm with 3 page frames.
![](https://www.geeksforgeeks.org/wp-content/uploads/LRU2.png)



