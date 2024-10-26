# SearchHouse
This is an entire full-stack search engine I'm creating and open-sourcing.

## Spider Features
Web crawlers are dime-a-dozen. I built this spider specifically to retrieve web pages
at a high volume as the beginning of my search engine pipeline.
- Spawn an arbitrary amount of routines to crawl pages
  - Each routine has its own delay to honor politeness
- Store frontier in SQLite database
  - Safe for concurrency
  - Crawler resumes where it left off if closed
  - Since the biggest bottleneck is IO (networking), retrieving frontier from disk is still viable
- Scheduler in main execution will allocate domains to each routine based on hash
  - Similar to a hashmap data structure
- Store pages, server responses, and timestamp as JSON object
- Hash URLs to store pages efficiently on disk

### Installation
```
git clone https://github.com/marceloclubhouse/searchHouse
cd searchHouse
go install searchHouse
go run main.go -h
```

### Running
```
go run main.go --seed="https://blog.marceloclub.house" --numRoutines=100
```

## License
This project is available under the GPL v3 license, see `LICENSE.txt` for more information.