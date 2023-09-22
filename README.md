# Distributed and Parrallel Programming with Go: Conway's Game of Life

### Introduction

The British mathematician John Horton Conway devised a cellular automaton named ‘The Game of Life’. The game resides on a 2-valued 2D matrix, i.e. a binary image, where the cells can either be ‘alive’ (pixel value 255 - white) or ‘dead’ (pixel value 0 - black). The game evolution is determined by its initial state and requires no further input. Every cell interacts with its eight neighbour pixels: cells that are horizontally, vertically, or diagonally adjacent. At each matrix update in time the following transitions may occur to create the next evolution of the domain:

- any live cell with fewer than two live neighbours dies
- any live cell with two or three live neighbours is unaffected
- any live cell with more than three live neighbours dies
- any dead cell with exactly three live neighbours becomes alive

Consider the image to be on a closed domain (pixels on the top row are connected to pixels at the bottom row, pixels on the right are connected to pixels on the left and vice versa). A user can only interact with the Game of Life by creating an initial configuration and observing how it evolves. Note that evolving such complex, deterministic systems is an important application of scientific computing, often making use of parallel architectures and concurrent programs running on large computing farms.

The task is to design and implement programs which simulate the Game of Life on an image matrix.

## Parallel Implementation
The parallel solution used multiple worker Go routines on a single machine to divide the workload between concurrent processes. This was completed with a SDL live viewer that visualised the world matrix for every step of the game of life. The program had to pass a number of tests to ensure that our implementation worked without error.

![Step 5](content/cw_diagrams-Parallel_5.png)

Parrallelising the workload greatly reduced the runtime of the program. Workers had to communicated between one another as they shared information so it was important to ensure there were no race conditions.

## Distributed Implementation

This implementation uses a number of AWS nodes to cooperativley calculate the new state of the game of life board and communicate the state between machines over a network.

![Step 5](content/cw_diagrams-Distributed_5.png)

The system is designed with the scalability of the AWS nodes in mind. Communication between the nodes was made to be efficient as possible. This was done with halo exchange.

### Halo Exchange

Halo exchange reduced the amount of data sent between the AWS nodes by only sending the edges of each nodes section of the matrix.

![Extension 1](content/cw_diagrams-Extensions_1.png)

## Comparing Solutions

The different solutions were benchmarked and then the results compared in a graph to find which solution was most efficient.

### Parrallelisation

We benchmarked our parallel implementation for 1000 turns
with up to 16 threads. Our graph shows how the run time
decreased as the computation was split to be concurrently
worked on by the threads. The greatest decrease can be seen
in the change in time per game between 1 and 2 threads,
where the workload is halved, Ideally, the runtime would be
consistently halved as the number of workers were doubled,
however this effect is diminishing and the optimisation of
our run time develops at a decreasing rate as communication
overhead and other inefficiencies, such as serially loading and
producing the PGM files - included in the above timings,
become more prevalent.

>>>INCL CHANNEL SOLUTION PHOTO.

### Distributed System

For our distributed system we benchmarked with up to 8
workers for 100 turns to account for the increase in runtime
when using AWS nodes. When run on a local network, our
distributed system produces results similar to our parallel
implementation. The division of work between components
reduced the runtime as the number of workers increased.
However there are some latency issues causing our results
to not follow a purely decreasing trend. The full effect of
latency and other pitfalls of a distributed system become
much more prevalent when we implemented our system with
AWS nodes.

When operating our distributed system over several AWS
nodes the impact of communication overhead becomes clear.
Repeatedly we produced results showing the runtime of our
system increasing at a steady rate as the number of AWS
workers increased. We believe this is because as we added
workers, the increase of consistent RPC calls between the
broker and the workers compounded issues caused by latency
and faults in the transmission of RPC request and response
data, as the broker communicated with the AWS nodes for
each turn of GOL. The inverse trends and great difference
between runtimes shown in figure 3 indicate the need for
further improvements on the system.

Some of these optimisations include:
- Reduce information being sent between the broker and AWS nodes each turn.
- Use multiple machines to take advantage of their hardware but have them connected on a local network or wired connection.

### Halo Exchange

Halo exchange optimised our AWS distributed system to a
fraction of the runtime. There is a decrease in runtime between
1 and 2 AWS worker nodes, which align with our expectations
for a distributed implementation, suggesting that the amount of
information sent between the nodes and the broker was greatly
impairing our system. However as the number of workers
increased, we began to see a small increase in the runtime
indicating that there is still more improvements that can be
made on the system or that using AWS nodes isn’t the most
efficient way of implementing a version of GOL that requires
great amounts of information sharing between concurrently
working components.

### Conclusion

Through our findings we have realised the opportunity to
optimise problems that can be parallelised by splitting work
between concurrently run procedures. In our parallel channel
implementation we witnessed a decreasing downward trend
in the runtime as we increased the number of threads used.
Other imporovements were made on the implementation such
as ensuring an even splitting of the work done by each spread.
Nonetheless, the limitation of a single machine’s hardware,
and the reliance on hyper-threading as we increased the
number of threads, prompted the necessity for a distributed
system. However, whilst operating efficiently on a single local
network, the expansion of our distributed system to multiple
AWS nodes greatly increased the runtime. The effects of communication overhead were clearly present in our result and,
despite the optimisations of halo exchange: greatly reducing
the information sent between AWS nodes and the broker,
suggested some weaknesses in the design of our distributed system with respect to the iterative nature of GOL.

