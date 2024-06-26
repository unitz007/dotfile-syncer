package main

type Commit struct {
	Id string `json:"id"`
}

type GitPull struct {
	Commits    []Commit `json:"commits"`
	HeadCommit Commit   `json:"head_commit"`
}
