# posrare
A tool for security researchers and bug bounty hunters, designed to find and prioritize URLs based on the uniqueness and entropy of words in a specified position in the URL path.

---

## Installation
`go install github.com/raverrr/posrare@latest` 

---

## Usage
Use `posrare -h` to see in-tool help

To use Posrare, simply pipe in your list of URLs. Use the `-p` flag to specify the position in the URL path you're interested in.
The position in the url is as seen here:

`https://example.com/position1/position2/position3/etc...`

Example sorting by position 1:

`cat urls.txt | posrare -p 1`


You can control the entropy level of the words with the `-e` flag. Higher entropy means more randomness. Lower entropy means more structure. Adjust this to suit your research needs:

`cat urls.txt | posrare -p 1 -e 3.5`


If you want to get a specific number of results, use the `-x` flag. This is best used with the `-v` flag when testing to figure out what entropy level to set:

`cat urls.txt | posrare -p 1 -e 3.5 -x 10 -v`


The output gives you URLs that contain unique words at the specified position. 

---

