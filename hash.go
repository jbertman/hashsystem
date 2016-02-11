package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"encoding/hex"
	"time"
	"github.com/cheggaaa/pb"
	"bufio"
	"runtime"
	"flag"
)

type res struct {
	path string
	hash string
	error string
}

func readLines (path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Worker Function
// Receive work on the jobs channel (file path)
// Send hash to the results channel
func worker (jobs <-chan string, results chan<- res) {
	// Sit on the channel
	for path := range jobs {
		// Open the file and compute the hash
		f, err := os.Open(path) // Can't do a defer because that will wait until the end
		if err != nil {
			results <- res{path: path, hash: "", error: err.Error()}
		}

		// We have a handle - copy the bytes to a buffer and compute the hash
		hash := md5.New()
		if _, err := io.Copy(hash, f); err != nil {
			results <- res{path: path, hash: "", error: err.Error()}
		}
		results <- res{path: path, hash: hex.EncodeToString(hash.Sum(nil)), error: ""}
		f.Close()
	}
}

func exists (path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func stringInSlice (a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

// Request index of all files, store it in a slice
// Create a worker pool, targeted at worker
// Throw jobs and listen for results
func main() {
	// Declare flags
	inputPtr := flag.String("i", "", "Input file (hashes - one per line)")
	outputPtr := flag.String("o", "", "Output directory")
	flag.Parse()
	
	if len(os.Args) != 5 {
		fmt.Println("[-] Error: Parameters -i (input file) and -o (output file) are both required.")
		flag.Usage()
		os.Exit(1)
	} else {
		if b, err := exists(*inputPtr); b == false {
			panic(err.Error())
		} else if b, err := exists(*outputPtr); b == false {
			panic(err.Error())
		} 
	}

	// Detect the platform at runtime. We should use a SystemEnv for Windows in case it's not C:\
	var platform, root string
	platform = runtime.GOOS
	fmt.Printf("[+] Detected %s as the operating system.\n", platform)
	if platform == "windows" {
		root = "C:\\"
	} else {
		root = "/"
	}

	// You're on the clock
	start := time.Now()

	// Read in the hashes
	fmt.Println("[+] Reading in the hashes...")
	hashes, err := readLines(*inputPtr)
	if err != nil {
		panic(err.Error())
	}

	// Create a list of paths
	things := make([]string, 0)

	// Walk some directory here
	// Put all the file paths in the things slice
	fmt.Println("[+] Walking files...")
	filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		// Files are statable, not directories, and at most 100 MB
		if err == nil {
			if f.Size() > 0 && f.Size() <= 100000000 && f.IsDir() == false {
				things = append(things, path)
			}
		}
		return nil
	})

	elapsed := time.Since(start)
	fmt.Println("[+] Index time was: ", elapsed)
	fmt.Printf("[+] There are %d items in the list.\n", len(things))
	fmt.Println("[+] Starting workers...")

	jobs := make(chan string, len(things)) // buffer is the length of the things slice (number of files)
	results := make(chan res, len(things)) // ^^

	// Start up some number of workers
	for w := 0; w < 5; w++ {
		go worker(jobs, results)
	}

	// We could rate-limit here using a ticker
	// limiter := time.Tick(time.Millisecond * 200)
	// and put <-limiter in the for loop below
	// Put all the file paths in the jobs channel
	for i := 0; i < len(things); i++ {
		jobs <- things[i]
	}
	// Close the channel
	close(jobs)

	//Create the results file
	f, err := os.Create("results.txt")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	// Progressbar
	bar := pb.StartNew(len(things))
	bar.ShowTimeLeft = false
	bar.SetMaxWidth(80)
	// We have to block for all possible jobs we are going to receive results for
	for i := 0; i < len(things); i++ {
		//fmt.Println(<-results)
		r := <-results
		bar.Increment()

		// Counter-intuitive multi-writes because they're actually more efficient
		if r.error != "" {
			f.WriteString("ERROR - ")
			f.WriteString(r.path)
			f.WriteString(" - ")
			f.WriteString(r.error)
			f.WriteString("\n")
		} else {
			if b := stringInSlice(r.hash, hashes); b == true {
				f.WriteString("MATCH - ")
				f.WriteString(r.path)
				f.WriteString(" - ")
				f.WriteString(r.hash)
				f.WriteString("\n")
			}
		}
		f.Sync()

	}
	elapsed = time.Since(start)
	time.Sleep(time.Millisecond * 201) // Let the progressbar refresh
	fmt.Println("[+] Processing complete! Total time: ", elapsed)

}