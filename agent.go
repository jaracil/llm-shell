package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/tmc/langchaingo/llms"
)

type Agent struct {
	Debug        bool
	Logic        string
	Llm          llms.Model
	Ctx          context.Context
	LastCommands []Command
	Provider     string
	Model        string
	APIKey       string
	WorkingDir   string
}

func (a *Agent) Command(command string) (*LlmOutput, error) {
	for iteration := 0; ; iteration++ {
		input := LlmInput{
			UserImput:    command,
			Directory:    a.WorkingDir,
			LastCommands: a.LastCommands,
		}
		if a.Debug {
			// Dump the status map in pretty printed JSON
			b, err := json.MarshalIndent(input, "", "  ")
			if err != nil {
				return nil, err
			}
			fmt.Printf("Status:\n%s\n\n", string(b))
		}

		llmInput := fmt.Sprintf("%s\n%s\n\n", a.Logic, input.String())
		// fmt.Println(llmInput)
		completion, err := llms.GenerateFromSinglePrompt(a.Ctx, a.Llm, llmInput,
			llms.WithTemperature(0.0), //0.7 is the default
		)
		if err != nil {
			return nil, err
		}

		// Isolate map from JSON response, removing leading and trailing no JSON characters
		completion = completion[strings.Index(completion, "{") : strings.LastIndex(completion, "}")+1]

		// fmt.Println(completion)
		var response LlmOutput
		err = json.Unmarshal([]byte(completion), &response)
		if err != nil {
			return nil, err
		}
		if a.Debug {
			// Dump the response map in pretty printed JSON
			b, err := json.MarshalIndent(response, "", "  ")
			if err != nil {
				return nil, err
			}
			fmt.Printf("Response:\n%s\n\n", string(b))
		}
		if response.Reasoning != "" {
			color.Green("Reasoning: %s\n", response.Reasoning)
		}
		if response.Command != "" {
			fmt.Printf("PWD: %s\n", a.WorkingDir)
			fmt.Println(response.Command)
		}
		cmd := exec.Command("bash", "-c", response.Command)
		if response.Directory != "" {
			cmd.Dir = response.Directory
			a.WorkingDir = response.Directory
		} else {
			cmd.Dir = a.WorkingDir
		}
		stdout, err := cmd.Output()
		exitCode := cmd.ProcessState.ExitCode()
		lastOutput := ""
		if err != nil {
			lastOutput = err.Error()
		} else {
			lastOutput = string(stdout)
		}
		truncated := false
		if len(lastOutput) > 2048 {
			// get the last 2048 characters
			lastOutput = lastOutput[len(lastOutput)-2048:]
			truncated = true
		}
		a.LastCommands = append(a.LastCommands, Command{
			Command:    response.Command,
			Directory:  response.Directory,
			Output:     lastOutput,
			ExitCode:   exitCode,
			UserOutput: response.UserOutput,
			UserInput:  command,
			Reasoning:  response.Reasoning,
			Iteration:  iteration,
			Truncated:  truncated,
		})
		// keep last 10 commands
		if len(a.LastCommands) > 10 {
			a.LastCommands = a.LastCommands[1:]
		}
		if response.UserOutput != "" {
			if response.Command != "" {
				response.UserOutput = fmt.Sprintf("%s\n%s", response.UserOutput, lastOutput)
			}
			return &response, nil
		}
		if iteration > 30 {
			return nil, fmt.Errorf("too many iterations")
		}
	}
}
