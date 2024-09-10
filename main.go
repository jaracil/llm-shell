package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chzyer/readline"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

func Repl(agent *Agent) {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		command, err := rl.Readline()
		if err != nil {
			log.Fatal(err)
		}
		if command == "" {
			continue
		}
		if command == "." {
			agent.LastCommands = []Command{}
			continue
		}
		if command == ".." {
			for _, c := range agent.LastCommands {
				s, _ := json.MarshalIndent(c, "", "  ")
				fmt.Printf("%s\n", s)
			}
			continue
		}
		if command == "/debug" {
			agent.Debug = true
			continue
		}
		if command == "/nodebug" {
			agent.Debug = false
			continue
		}
		r, err := agent.Command(command)
		if err != nil {
			fmt.Printf("internal error: %s\n", err)
			continue
		}
		if r.UserOutput != "" {
			fmt.Println(r.UserOutput)
		}

	}
}

func main() {
	// Define command line params
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	var provider string
	flag.StringVar(&provider, "provider", "openai", "ai engine to use (openai,google,)")
	var model string
	flag.StringVar(&model, "model", "", "ai model to use (gpt-3.5-turbo,...)")
	var apiKey string
	flag.StringVar(&apiKey, "apikey", "", "api key for the ai provider")
	flag.Parse()

	ctx := context.Background()
	var llm llms.Model
	var err error
	switch provider {
	case "openai":
		if model == "" {
			model = "gpt-4o-mini"
		}
		llm, err = openai.New(openai.WithToken(apiKey), openai.WithModel(model))
	case "google":
		if model == "" {
			model = "gemini-1.5-flash"
		}
		llm, err = googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel(model))
	}

	if err != nil {
		log.Fatal(err)
	}

	logic, err := os.ReadFile("logic.txt")
	if err != nil {
		log.Fatal(err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	agent := &Agent{
		Logic:        string(logic),
		Llm:          llm,
		Ctx:          ctx,
		Debug:        debug,
		Provider:     provider,
		Model:        model,
		APIKey:       apiKey,
		LastCommands: []Command{},
		WorkingDir:   pwd,
	}
	fmt.Printf("Using %s model %s\n", agent.Provider, agent.Model)
	go Repl(agent)

	<-ctx.Done()
}
