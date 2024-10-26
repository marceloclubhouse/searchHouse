package spider

type distributor struct {
	numThreads       int
	frontier         Frontier
	workingDirectory string
}
