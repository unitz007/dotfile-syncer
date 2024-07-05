package main

type Commit struct {
	Id   string `json:"id"`
	Time string `json:"commit_time"`
}

type GitWebHookCommitResponse struct {
	HeadCommit struct {
		Id string `json:"id"`
	} `json:"head_commit"`
}

type GitPullTransform struct {
	IsSync       bool    `json:"is_synced"`
	RemoteCommit *Commit `json:"remote_commit"`
	LocalCommit  *Commit `json:"local_commit"`
}

type GitHttpCommitResponse struct {
	Sha    string `json:"sha"`
	Commit struct {
		Name string `json:"name"`
	} `json:"commit"`
}

func InitGitTransform(localCommit *Commit, remoteCommit *Commit) GitPullTransform {
	return GitPullTransform{
		LocalCommit:  localCommit,
		RemoteCommit: remoteCommit,
		IsSync: func() bool {
			if localCommit != nil && remoteCommit != nil {
				return localCommit.Id == remoteCommit.Id
			}
			return false
		}(),
	}
}
