# Proof of Work (PoW) DDOS protection TCP server 

## 1. Description
This project is a solution for the following problem statement.

Design and implement “Word of Wisdom” tcp server.
TCP server should be protected from DDOS attacks with the Prof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
 + The choice of the POW algorithm should be explained.
 + After Prof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.
 + Docker file should be provided both for the server and for the client that solves the POW challenge

![main workflow](https://raw.githubusercontent.com/klimenkoOleg/pow-ddos-protection/main/doc/main_workflow.png)

### 2.1 Requirements
+ [Go 1.20+](https://go.dev/dl/) installed (to run tests, start server or client without Docker)
+ [Docker](https://docs.docker.com/engine/install/) installed (to run docker-compose)

### 2.2 Start server and client by docker-compose:
```
docker compose up
```

This command will fire up docker containers for server and client, performs compilation, runs test, run the server and client services - inside the docker containers.

## 4. Protocol definition
This solution uses TCP-based protocol. 
Message is encoded by gob and then encrypted by an asymmetric encryption algorithm - RSA.
Message sent from client to server and back has universal format:
+ header - integer number to indicate, which type of request was sent (analogue of URL in HTTP-protocol)
+ payload - array of bytes to handle any kind of meaingful data (depends on type of request)

### 4.1 Types of requests
Solution supports 5 types of requests, switching by header:
+ 0 - 	Step1ChallengeRequest            // from client to server - request new challenge from server
+ 1 - 	Step2TwoChallengeResponse        // from server to client - message with challenge for client
+ 2	-  Step3QuoteRequest                // from client to server - message with solved challenge
+ 3 - 	Step4QuoteResponse               // from server to client - message with useful info is solution is correct, or with error if not

## 5. Proof of Work
Idea of Proof of Work for DDOS protection is that client, which wants to get some resource from server, 
should firstly solve some challenge from server. 
This challenge should require more computational work on client side and verification of challenge's solution - much less on the server side.

### 5.1 Selection of an algorithm
There is some different algorithms of Proof Work. 
I compared next three algorithms as more understandable and having most extensive documentation:
+ [Proof of Usewful work](https://wiki.internetcomputer.org/wiki/Proof_of_Useful_Work)
+ [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree)
+ [Hashcash](https://en.wikipedia.org/wiki/Hashcash)
+ [Guided tour puzzle](https://en.wikipedia.org/wiki/Guided_tour_puzzle_protocol)

#### PoW algorithm I'd like to use
I'd like to use Proof of Useful work because all other PoW algorithms use computer resources without signiificant outcome.
The 

#### PoW algorithm I selected on practice at the moment (mid 2023)
After comparison, I chose Hashcash. Other algorithms have next disadvantages:
+ In Merkle tree server should do too much work to validate client's solution. For tree consists of 4 leaves and 3 depth server will spend 3 hash calculations.
+ In guided tour puzzle client should regularly request server about next parts of guide, that complicates logic of protocol.

Hashcash, instead has next advantages:
+ simplicity of implementation
+ lots of documentation and articles with description
+ simplicity of validation on server side
+ possibility to dynamically manage complexity for client by changing required leading zeros count

Of course Hashcash also has disadvantages like:

1. Compute time depends on power of client's machine. 
For example, very weak clients possibly could not solve challenge, or too powerful computers could implement DDOS-attackls.
But complexity of challenge could be dynamically solved by changing of required zeros could from server.
2. Pre-computing challenges in advance before DDOS-attack. 
Some clients could parse protocol and compute many challenges to apply all of it in one moment.
It could be solved by additional validation of hashcash's params on server. 
For example, on creating challenge server could save **rand** value to Redis cache and check it's existence on verify step
(that is implemented in this solution).

But all of those disadvantages could be solved in real production environment. 


## Proof of Usewful work references:
+ Proof of Usewful work. Research paper. https://github.com/cfschilham/kophos/blob/master/proof_of_useful_work.pdf
+ A Proof of Useful Work for Artificial Intelligence on the Blockchain https://arxiv.org/pdf/2001.09244.pdf
+ Coin.AI: A Proof-of-Useful-Work Scheme for Blockchain-Based Distributed Deep Learning https://www.mdpi.com/1099-4300/21/8/723


