package main

import "encoding/json"

type Command struct {
	Command    string `json:"command"`
	Directory  string `json:"directory"`
	Output     string `json:"output"`
	ExitCode   int    `json:"exitCode"`
	UserInput  string `json:"userInput"`
	UserOutput string `json:"userOutput"`
	Reasoning  string `json:"reasoning"`
	Iteration  int    `json:"iteration"`
	Truncated  bool   `json:"truncated"`
}

type LlmInput struct {
	UserImput    string `json:"userInput"`
	Directory    string `json:"directory"`
	LastCommands []Command
}

func (l *LlmInput) String() string {
	b, _ := json.Marshal(l)
	return string(b)
}

type LlmOutput struct {
	UserOutput string `json:"userOutput"`
	Command    string `json:"command"`
	Directory  string `json:"directory"`
	Reasoning  string `json:"reasoning"`
}
