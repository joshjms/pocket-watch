package models

type Request struct {
	Code     string   `json:"code"`
	Language string   `json:"language"`
	Stdin    []string `json:"stdin"`
}

type Response struct {
	Verdict []string `json:"verdict"`
	Stdout  []string `json:"stdout"`
	Stderr  []string `json:"stderr"`
	Time    []string `json:"time"`
	Memory  []string `json:"memory"`
}
