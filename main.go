package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/krognol/go-wolfram"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

var wolframClient *wolfram.Client

func printCommandEvents(channel <-chan *slacker.CommandEvent) {
	for event := range channel {
		fmt.Println("Command Events :")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func main() {
	godotenv.Load(".env")
	bot := slacker.NewClient(os.Getenv("Slack_Bot_Token"), os.Getenv("Slack_App_Token"))
	client := witai.NewClient(os.Getenv("Wit_AI_Server_Token"))
	wolframClient = &wolfram.Client{AppID: os.Getenv("Wolfram_App_Id")}
	go printCommandEvents(bot.CommandEvents())
	bot.Command("query - <message>", &slacker.CommandDefinition{
		Description: "send any question as query",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			query := request.Param("message")
			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			data, _ := json.MarshalIndent(msg, "", "    ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				fmt.Println("There is an error")
			}
			response.Reply(res)
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
