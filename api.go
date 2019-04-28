package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/nlopes/slack"
	"golang.org/x/oauth2"
)

type Package struct {
	FullName      string
	Description   string
	StarsCount    int
	ForksCount    int
	LastUpdatedBy string
}
type ListOptions struct {
	// For paginated result sets, page of results to retrieve.
	Page int `url:"page,omitempty"`
	// For paginated result sets, the number of results to include per page.
	PerPage int `url:"per_page,omitempty"`
}
type Commit struct {
	Message string `json:"Message"`
}

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable " + name)
	}
	return v
}
func main() {
	context := context.Background()
	tokenService := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "voir notes"},
	)
	tokenClient := oauth2.NewClient(context, tokenService)

	client := github.NewClient(tokenClient)

	repo, _, err := client.Repositories.Get(context, "s4mking", "Gazooy")

	if err != nil {
		fmt.Printf("Problem in getting repository information %v\n", err)
		os.Exit(1)
	}

	pack := &Package{
		FullName:    *repo.FullName,
		Description: *repo.Description,
		ForksCount:  *repo.ForksCount,
		StarsCount:  *repo.StargazersCount,
	}
	var commitInfo []*github.RepositoryCommit
	commitInfo, _, err = client.Repositories.ListCommits(context, "s4mking", "Gazooy", nil)

	if err != nil {
		fmt.Printf("Problem in commit information %v\n", err)
		os.Exit(1)
	}
	pack = pack
	var lastMessage = *commitInfo[0].Commit.Message
	var lastAutor = *commitInfo[0].Commit.Author.Name
	token := getenv("SLACKTOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	groups, err := api.GetGroups(false)

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, group := range groups {
		fmt.Printf("ID: %s, Name: %s\n", group.ID, group.Name)

	}
	//L'idée serait par la suite de faire des tests pour suivre l'avancée des projets sur github avec une liste de commandes à effectuer sur slack
	//A faire : détecter des commandes spécifique spas juste le test d'entrée classique de texte voir aussi sur l'api de github ce qu'il est possible dee rajouter
Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Print("Event Received: ")
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				fmt.Println("Connection counter:", ev.ConnectionCount)

			case *slack.MessageEvent:
				fmt.Printf("Message: %v\n", ev)
				// info := rtm.GetInfo()
				// prefix := fmt.Sprintf("<@%s> ", info.User.ID)
				// if ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
				rtm.SendMessage(rtm.NewOutgoingMessage(lastMessage+" "+lastAutor, ev.Channel))
				// }

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				//Take no action
			}
		}
	}
}
