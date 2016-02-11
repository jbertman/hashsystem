# hashsystem
Hash the entire system (all files under 100MB) looking for hashes in a list
##Dependencies
github.com/cheggaaa/pb

###Install pb using the built-in package manager (get)
    export GOPATH=/home/user/projects/
    go get github.com/cheggaaa/pb

##Usage
hashsystem takes in two required parameters:

-i (input file) is a text file containing a list of "known-bad" hashes. 

-o (output director) is a valid path where the program will output the results.txt file.


The main thread searches each hash result as it comes in from a worker. Matches have the following structure:

    MATCH - [path] - [hash]


Initially unstatable directories or files that the contextual user does not have access to will cause an error during the hashing process. Errors have the following structure:

    ERROR - [path] - [error]


##The results.txt file can then be searched for matches:

    grep MATCH results.txt


##Or errors:

    grep ERROR results.txt


##How many matches were there over the entire scan?

    grep MATCH results.txt | wc -l


etc.


It is common to see several thousand errors and can usually be safely disregarded.


##Things of note
The initial file index process caps the file size at 100MB in the interest of speed. That means any file over 100MB will not get hashed. This can be changed in the code in the anonymous function declared during the filewalk process. In the future, this will be configurable.


hashsystem can be very hard on a machine. It does NOT rate-limit the hashing. The rate-limiting ability is currently disabled (can be enabled in the code) but does not reflect adequate testing for speed (the rate-limit can be tuned). In the future, the rate-limit should be tuned dynamically to account for percentage of CPU. 


##You can run hashsystem locally using the go run command:

    go run hash.go -i hashes.txt -o ./


##Or you can build a small (~2MB) executable

    go build hash.go

    ./hash -i hashes.txt -o ./
