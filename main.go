package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	gobot "github.com/danrusei/gobot-bsky"
	"github.com/enescakir/emoji"
	"github.com/joho/godotenv"
)

var LAST_STATUS = ""

func main() {
	godotenv.Load()

	var BSKY_PDS = os.Getenv("BSKY_PDS")
	var BSKY_HANDLE = os.Getenv("BSKY_HANDLE")
	var BSKY_PASSWD = os.Getenv("BSKY_PASSWD")

	var DC_USER_ID = os.Getenv("DC_USER_ID")
	var DC_TOKEN = os.Getenv("DC_TOKEN")
	var DC_IGNORESUFFIX = os.Getenv("DC_IGNORESUFFIX")

	fmt.Printf("Connecting to bluesky with %s:%s\n", BSKY_PDS, BSKY_HANDLE)
	bsapi, err := NewBSAPI(BSKY_PDS, BSKY_HANDLE, BSKY_PASSWD)
	if err != nil {
		panic("Error connecting to the bluesky API: " + err.Error())
	}
	dcapi, err := NewDCAPI(DC_TOKEN)
	if err != nil {
		panic("Error connecting to the discord API: " + err.Error())
	}
	defer dcapi.discord.Close()

	dcapi.discord.Identify.Intents = discordgo.IntentsGuildPresences
	err = dcapi.discord.Open()
	if err != nil {
		panic("Error opening discord comunications: " + err.Error())
	}

	dcapi.discord.AddHandler(func(s *discordgo.Session, p *discordgo.PresenceUpdate) {
		// dbjson, _ := json.Marshal(p)
		// fmt.Printf("%s\n", dbjson)
		if p.User.ID != DC_USER_ID {
			return
		}
		if len(p.Activities) == 0 {
			return
		}

		for _, activity := range p.Activities {
			if activity.Type != discordgo.ActivityTypeCustom {
				continue
			}
			v := activity.State
			if LAST_STATUS != v {
				LAST_STATUS = v
				fmt.Printf("User's custom status: %s\n", v)
				if strings.HasSuffix(v, DC_IGNORESUFFIX) {
					fmt.Printf("IGNORING\n")
					return
				}
				moji := emoji.Parse(activity.Emoji.Name)
				var mojitext string
				if !strings.Contains(moji, ":") && activity.Emoji.Name != "" {
					mojitext = moji + ": "
				}
				fmt.Println(moji)
				fmt.Println(mojitext)
				bsapi.DoPost(fmt.Sprintf("%s%s\n-Discord status", mojitext, v))
				return
			}
		}
	})

	fmt.Println("Initialised")
	if gobot.ErrExpiredToken.Error() != "" {
		fmt.Println("it did actually update")
	}
	select {}
}

type DCAPI struct {
	discord *discordgo.Session
}

func NewDCAPI(token string) (DCAPI, error) {
	fmt.Println("Connecting to the discord API")
	discord, err := discordgo.New(token)
	if err != nil {
		return DCAPI{}, err
	}
	return DCAPI{discord: discord}, err
}

type BSAPI struct {
	Agent gobot.BskyAgent
	CTX   context.Context
}

func NewBSAPI(pds string, handle string, passwd string) (BSAPI, error) {
	ctx := context.Background()

	agent := gobot.NewAgent(ctx, pds, handle, passwd)
	err := agent.Connect(ctx)
	if err != nil {
		return BSAPI{}, err
	}
	return BSAPI{Agent: agent, CTX: ctx}, nil
}

func (a *BSAPI) DoPost(content string) (string, string, error) {
	post, err := gobot.NewPostBuilder(content).
		Build()
	if err != nil {
		fmt.Printf("Got error: %v\n", err)
		return "", "", err
	}

	cid, uri, err := a.Agent.PostToFeed(a.CTX, post)
	if err != nil {
		fmt.Printf("Got error: %v\n", err)
		return "", "", err
	} else {
		fmt.Printf("Succes: Cid = %v , Uri = %v\n", cid, uri)
		return cid, uri, nil
	}
}
