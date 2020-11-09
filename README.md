# CSA Coursework: Game of Life

**IMPORTANT: Monday 16.11 at 10:00 - COURSEWORK Q&A on Teams**

This is the Computer Systems A summative coursework. The coursework is worth 80% of the unit mark. It is to be completed in your programming pairs. You must report any change to your pairing to the unit director *before* starting your assignment. It runs over 5 weeks and the deadline for submitting all your work is **Friday 11 December 13:00**.

Talk to each other regularly and make sure you manage your team well. Let us know about issues before they grow to affect your team’s performance.

**Do not plagiarise.** Both team members should understand all code developed in detail.

## Task Overview

### Introduction

The British mathematician John Horton Conway devised a cellular automaton named ‘The Game of Life’. The game resides on a 2-valued 2D matrix, i.e. a binary image, where the cells can either be ‘alive’ (pixel value 255 - white) or ‘dead’ (pixel value 0 - black). The game evolution is determined by its initial state and requires no further input. Every cell interacts with its eight neighbour pixels: cells that are horizontally, vertically, or diagonally adjacent. At each matrix update in time the following transitions may occur to create the next evolution of the domain:

- any live cell with fewer than two live neighbours dies
- any live cell with two or three live neighbours is unaffected
- any live cell with more than three live neighbours dies
- any dead cell with exactly three live neighbours becomes alive

Consider the image to be on a closed domain (pixels on the top row are connected to pixels at the bottom row, pixels on the right are connected to pixels on the left and vice versa). A user can only interact with the Game of Life by creating an initial configuration and observing how it evolves. Note that evolving such complex, deterministic systems is an important application of scientific computing, often making use of parallel architectures and concurrent programs running on large computing farms.

Your task is to design and implement programs which simulate the Game of Life on an image matrix.

### Skeleton Code

To help you along, you are given a simple skeleton project. The skeleton includes tests and an SDL-based visualiser. All parts of the skeleton are commented. All the code has been written in Go. You will not be required to write any C code. If you have any questions about the skeleton please ask a TA for help.

You **must not** modify any of the files ending in `_test.go`. We will be using these tests to judge the correctness of your implementation.

### Submission

The coursework requires two independent implementations. You will be required to submit **both** implementations (assuming both were attempted). Every student is required to upload their full work to Blackboard. There will be three sperate submissions points on Blackboard - one for the report and two for each implementation.

For the report you must submit a single file called `report.pdf`.

For the parallel implementation you must submit a single zip file called `parallel.zip`. It must contain all the code required to compile and run the program.

For the distributed implementation you must submit a single zip file called `distributed.zip`. It must contain all the code required to compile and run the program.

Submitting different filenames or file formats (e.g. `.docx`, `.tex`, `.7z` or `.rar`) will result in a mark penalty.

Make sure you submit it early (not last minute!) to avoid upload problems. **Each team member has to upload an identical copy of the team's work.**

## Stage 1 - Parallel Implementation

In this stage you are required to write code to evolve Game of Life using multiple worker goroutines on a single machine. Below are some suggested steps to help you get started. You are *not* required to follow them. Your implementation will be marked against the success criteria outlined below.

### Step 1

Implement the Game of Life logic as it was described in the task introduction. We suggest starting with a single-threaded implementation that will serve as a starting point in subsequent steps. Your Game of Life should evolve for the number of turns specified in `gol.Params.Turns`.

Test your serial, single-threaded code using `go test -v -run=TestGol/-1$`. All the tests ran should pass.

### Step 2

Parallelise your Game of Life so that it uses worker threads to calculate the new state of the board. You should implement a distributor that tasks different worker threads to operate on different parts of the image in parallel. The number of worker threads you should create is specified in `gol.Params.Threads`.

*Note: You are free to design your system as you see fit, however, we encourage you to primarily use channels*

Test your code using `go test -v -run=TestGol`. You can use tracing to verify the correct number of workers was used this time.

### Step 3

The lab sheets included the use of a timer. Now using a ticker, report the number of cells that are still alive *every 2 seconds*. To report the count use the `AliveCellsCount` event.

Test your code using `go test -v -run=TestAlive`.

### Step 4

Implement logic to output the state of the board after all turns have completed as a PGM image.

Test your code using `go test -v -run=TestPgm`. Finally, run `go test -v` and make sure all tests are passing.

### Step 5

Implement logic to visualise the state of the game using SDL. Also implement the following control rules. Note that the goroutine running SDL provides you with a channel containing the relevant keypresses.

- If `s` is pressed, generate a PGM file with the current state of the board.
- If `q` is pressed, generate a PGM file with the current state of the board and then terminate the program. Your program should *not* continue to execute all turns set in `gol.Params.Turns`.
- If `p` is pressed, pause the processing and print the current turn that is being processed. If `p` is pressed again resume the processing and print `"Continuing"`. It is *not* necessary for `q` and `s` to work while the execution is pauesed.

Test the visualisation and control rules by running `go run .`

### Success Criteria

- Pass all test cases under `TestGol`, `TestAlive` and `TestPgm`.
- Use the correct number of workers as requested in `gol.Params`.
- Display the live progress of the game using SDL.
- Ensure that all keyboard control rules work correctly.
- Use benchmarks to measure the performance of your parallel program.
- The implementation must be free of deadlocks and race conditions.

### In your Report

- Discuss the goroutines you used and how they work together.
- Explain and analyse the benchmark results obtained.
- Analyse how your implementation scales as more workers are added.
- Briefly discuss your methodology for aquiring any results or measurements.

## Stage 2 - Distributed Implementation

In this stage you are required to create an implementation that uses a number of AWS nodes to calculate the new state of the board.

[TODO] Steps?
- Single threaded, distributed
- Multi threaded, distributed
- Extension

[TODO] Use a locally running client to connect to the distributed system?

[TODO] Some sort of fault tolerance?

### Success Criteria

- Pass all tests.
- Output the correct PGM images.
- Ensure the keyboard control rules work as needed.
- Use benchmarks to measure the performance of your distributed program.

*There is __no need__ to display the live progress of the game using SDL. However, you will still need to run a blank SDL window to register the keypresses.*

### In your report

- Discuss the system design and reasons for any decisions made.
- Explain and analyse the benchmark results obtained.
- Briefly discuss your methodology for aquiring benchmark results.

## Extensions

[TODO]

-----------------------------------------------------------------------

## Mark Scheme

[TODO]

## Report

You need to submit a CONCISE (strictly max 6 pages) report which should cover the following topics:

Functionality and Design: Outline what functionality you have implemented, which problems you have solved with your implementations and how your program is designed to solve the problems efficiently and effectively.

Critical Analysis: Describe briefly the other experiments and analysis you carried out, provide a selection of appropriate results. Keep a history of your implementations and provide benchmark results from various stages. Explain and analyse the benchmark results obtained. Analyse the important factors responsible for virtues and limitations of your implementations.

Make sure your team’s names and user names appear on page 1 of the report. Don't include a cover page.

## Viva

You will be required to demostrate your implementations in the viva. This will include running tests as well as showing PGM image output and working keyboard control.

As part of the viva, we will also discuss your report. You should be prepared to discuss and expand on any points mentioned in your report.

All vivas will run in Week 11.

## Workload and Time Management

It is important to carefully manage your time for this assignment. Do not spend hours trying to debug on your own; use pair programming, seek help from our teaching assistants during scheduled labs and ask questions on Teams.
