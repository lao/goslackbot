package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// func main() {

// 	// Load Env variables from .dot file
// 	godotenv.Load(".env")

// 	appToken := os.Getenv("SLACK_APP_TOKEN")
// 	token := os.Getenv("SLACK_BOT_TOKEN")

// 	// Create a new client to slack by giving token
// 	// Set debug to true while developing
// 	// Also add a ApplicationToken option to the client
// 	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

// 	// go-slack comes with a SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
// 	socketClient := socketmode.New(
// 		client,
// 		socketmode.OptionDebug(true),
// 		// Option to set a custom logger
// 		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
// 	)

// 	// Create a context that can be used to cancel goroutine
// 	ctx, cancel := context.WithCancel(context.Background())
// 	// Make this cancel called properly in a real program , graceful shutdown etc
// 	defer cancel()

// 	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
// 		// Create a for loop that selects either the context cancellation or the events incomming
// 		for {
// 			select {
// 			// inscase context cancel is called exit the goroutine
// 			case <-ctx.Done():
// 				log.Println("Shutting down socketmode listener")
// 				return
// 			case event := <-socketClient.Events:
// 				// We have a new Events, let's type switch the event
// 				// Add more use cases here if you want to listen to other events.
// 				switch event.Type {
// 				// handle EventAPI events
// 				case socketmode.EventTypeEventsAPI:
// 					// The Event sent on the channel is not the same as the EventAPI events so we need to type cast it
// 					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
// 					if !ok {
// 						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
// 						continue
// 					}

// 					// Now we have an Events API event, but this event type can in turn be many types, so we actually need another type switch
// 					log.Println(eventsAPIEvent)

// 					switch eventsAPIEvent.Type {
// 					case slackevents.CallbackEvent:
// 						// We need to send an Acknowledge to the slack server
// 						// socketClient.Ack(*event.Request)

// 						innerEvent := eventsAPIEvent.InnerEvent
// 						handleMessages(client, eventsAPIEvent, innerEvent)
// 						// We need to send an Acknowledge to the slack server
// 						socketClient.Ack(*event.Request)
// 					default:
// 						log.Printf("Unexpected Events API event received: %v\n", eventsAPIEvent)
// 						socketClient.Ack(*event.Request)
// 					}
// 				}
// 			}
// 		}
// 	}(ctx, client, socketClient)

// 	socketClient.Run()
// }

func main() {

	godotenv.Load(".env")

	token := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:

				switch event.Type {

				case socketmode.EventTypeEventsAPI:

					eventsAPI, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPI: %v\n", event)
						continue
					}

					socketClient.Ack(*event.Request)
					log.Println(eventsAPI)
					handleMessages(client, eventsAPI, eventsAPI.InnerEvent)
				}
			}
		}
	}(ctx, client, socketClient)

	socketClient.Run()
}

func logMessageEvent(ev *slackevents.MessageEvent) {
	log.Println(ev.Text)
	log.Println(ev.Files)
	log.Println(ev.User)
	log.Println(ev.Channel)
	log.Println(ev.BotID)
	log.Println(ev.SubType)
	log.Println(ev.ClientMsgID)
}

func isBotMessage(event slackevents.EventsAPIEvent) bool {
	// get event data
	data := event.InnerEvent.Data

	// type switch to get message event
	switch ev := data.(type) {
	case *slackevents.MessageEvent:
		// if bot id is not empty then it is a bot message
		if ev.BotID != "" {
			return true
		}
	case *slackevents.AppMentionEvent:
		if ev.BotID != "" {
			return true
		}
	case *slackevents.MessageMetadataPostedEvent:
		if ev.BotId != "" {
			return true
		}
	case *slackevents.MessageMetadataUpdatedEvent:
		if ev.BotId != "" {
			return true
		}
	case *slackevents.MessageMetadataDeletedEvent:
		if ev.BotId != "" {
			return true
		}
	default:
		return false
	}

	return false
}

func handleMessages(bot *slack.Client, event slackevents.EventsAPIEvent, innerEvent slackevents.EventsAPIInnerEvent) {

	// if event is from bot skip
	if isBotMessage(event) {
		return
	}

	// get innerEvent data message type, text and files
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		logMessageEvent(ev)
		message := ev.Text
		channel := ev.Channel
		user := ev.User

		handleMessage(message, channel, user, bot)
	case *slackevents.AppMentionEvent:
		log.Println(ev.Text)
		// log.Println(ev.Files)
		log.Println(ev.User)
		log.Println(ev.Channel)
		log.Println(ev.BotID)
		// log.Println(ev.SubType)
		// log.Println(ev.ClientMsgID)
		// get the message text
		message := ev.Text
		// get the channel id
		channel := ev.Channel
		// get the user id
		user := ev.User
		handleMessage(message, channel, user, bot)
	default:
		log.Println("Not a message event")
	}

}

func handleMessage(message string, channel string, user string, bot *slack.Client) {
	// check if the message matches the pattern
	if strings.Contains(message, "hello") {
		// send a message to the channel
		_, _, err := bot.PostMessage(channel, slack.MsgOptionText("Hello <@"+user+">", false))
		if err != nil {
			log.Println(err)
		}
	} else if strings.Contains(message, "transfer") {
		// get the file from the slack message
		// file, err := bot.GetFileInfo(event.File.ID)
		// if err != nil {
		// 	log.Println(err)
		// }
		_, _, err := bot.PostMessage(channel, slack.MsgOptionText("transfering...", false))
		if err != nil {
			log.Println(err)
		}
		// send a message to the channel
	} else if strings.Contains(message, "help") {
		// send a message to the channel
		_, _, err := bot.PostMessage(channel, slack.MsgOptionText("I can help you", false))
		if err != nil {
			log.Println(err)
		}
	} else {
		// send a message to the channel
		_, _, err := bot.PostMessage(channel, slack.MsgOptionText("I don't understand", false))
		if err != nil {
			log.Println(err)
		}
	}
}
