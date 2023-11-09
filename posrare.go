package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type Job struct {
	word string
	url  string
}

var entropyCache = make(map[string]float64)
var cacheMutex sync.Mutex

func calculateEntropy(word string) float64 {
	cacheMutex.Lock()
	if entropy, ok := entropyCache[word]; ok {
		cacheMutex.Unlock()
		return entropy
	}
	cacheMutex.Unlock()

	m := make(map[rune]float64)

	// character freq
	for _, c := range word {
		m[c]++
	}

	//entropy
	entropy := 0.0
	for _, freq := range m {
		freq /= float64(len(word))
		entropy += freq * math.Log2(freq)
	}

	cacheMutex.Lock()
	entropyCache[word] = -entropy
	cacheMutex.Unlock()

	return -entropy
}

func worker(jobs <-chan Job, wordFrequency map[string]int, urlWord map[string]string, wg *sync.WaitGroup, mutex *sync.Mutex) {
	defer wg.Done()
	for job := range jobs {
		mutex.Lock()
		wordFrequency[job.word]++
		// Store the url AND word in map
		urlWord[job.word] = job.url
		mutex.Unlock()
	}
}

func main() {
	position := flag.Int("p", 1, "Position in the URL path to extract word from")
	topX := flag.Int("x", -1, "Number of URLs to return that contain the least common words at the specified position. If not specified or if set to -1, all URLs will be returned. This flag is best used with the -v flag when testing to determine what entropy level to set.")
	entropyLevel := flag.Float64("e", 3.5, "Maximum entropy level for the word at the specified position. Words with higher entropy will be ignored. Default is 3.5.")
	verbose := flag.Bool("v", false, "Enable verbose output. When enabled, additional information such as total unique words and entropy level will be displayed.")

	flag.Usage = func() {
		fmt.Println("Usage: posrare [OPTIONS]")
		fmt.Println("Reads URLs from stdin and outputs the ones where the word at the specified position (-p) in the URL path has an entropy below the specified level, and occur with the least frequency.")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	wordFrequency := make(map[string]int)
	urlWord := make(map[string]string)

	jobs := make(chan Job, 100)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	//workers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(jobs, wordFrequency, urlWord, &wg, &mutex)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		u, err := url.Parse(scanner.Text())
		if err != nil {
			continue
		}

		// path stuff
		path := u.Path
		pathArray := strings.Split(path, "/")
		if len(pathArray) > *position {
			word := pathArray[*position]
			// Check entropy
			if calculateEntropy(word) <= *entropyLevel {
				jobs <- Job{word: word, url: scanner.Text()}
			}
		}
	}
	close(jobs)

	wg.Wait()

	var topWords []string

	for word := range wordFrequency {
		topWords = append(topWords, word)
	}
	// sort-em
	sort.Slice(topWords, func(i, j int) bool {
		if wordFrequency[topWords[i]] != wordFrequency[topWords[j]] {
			return wordFrequency[topWords[i]] < wordFrequency[topWords[j]]
		}
		return topWords[i] < topWords[j]
	})

	// Not pretty....
	if *topX == -1 || *topX > len(topWords) {
		*topX = len(topWords)
	}
	topWords = topWords[:*topX]
	//sort.Strings(topWords)

	//results
	if *verbose {
		fmt.Printf("%s: %d\n", cyan("Total unique words at selected position"), len(wordFrequency))
		fmt.Printf("%s: %f\n", cyan("Entropy level"), *entropyLevel)
		fmt.Printf("%s:\n", yellow("Sample of "+fmt.Sprint(*topX)))
	}
	for _, word := range topWords {
		u, _ := url.Parse(urlWord[word])
		pathArray := strings.Split(u.Path, "/")
		queryString := ""
		if u.RawQuery != "" {
			queryString = "?" + u.RawQuery
		}
		highlightedURL := u.Scheme + "://" + u.Host + strings.Replace(u.Path, pathArray[*position], red(pathArray[*position]), 1) + queryString
		if *verbose {
			fmt.Printf("%s: ", cyan(fmt.Sprintf("(Entropy at position %d: %.2f)", *position, calculateEntropy(pathArray[*position]))))
		}
		fmt.Println(highlightedURL)
	}
}
