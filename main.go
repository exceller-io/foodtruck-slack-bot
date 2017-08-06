package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/rprakashg/foodtruck-slack-bot/seattlefoodtruck"
)

const s3Bucket = "https://s3-us-west-2.amazonaws.com/seattlefoodtruck-uploads-prod/"

func main() {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

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
				info := rtm.GetInfo()
				prefix := fmt.Sprintf("<@%s> ", info.User.ID)

				if ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
					respond(rtm, ev, prefix)
				}

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
func respond(rtm *slack.RTM, msg *slack.MessageEvent, prefix string) {
	text := msg.Text
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	if text == "help" {
		response := fmt.Sprintf("%s: \n", "*You can ask me*")
		response += fmt.Sprintf("%s \n", "• *show neighborhoods* - to see neighborhoods served")
		response += fmt.Sprintf("%s \n", "• *show locations in <neighborhood>* - to see food truck locations in a neighborhood")
		response += fmt.Sprintf("%s \n", "• *show trucks at <location>* - to see food trucks at a location")

		rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
	} else if text == "show neighborhoods" {
		showNeighborhoods(rtm, msg.Channel)
	} else if strings.Contains(text, "show locations") {
		showLocations(rtm, text, msg.Channel)
	} else if strings.Contains(text, "show trucks") {
		showTrucks(rtm, text, msg.Channel)
	} else {
		rtm.SendMessage(rtm.NewOutgoingMessage("Sorry I cannot help you with this, please try help to see things you can ask me", msg.Channel))
	}
}
func showNeighborhoods(rtm *slack.RTM, channel string) {
	var message string
	p, _ := seattlefoodtruck.NewProxy("https://www.seattlefoodtruck.com")
	resp, err := p.GetNeighborhoods()
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(err.Error(), channel))
		return
	}
	if len(resp.Neighborhoods) == 0 {
		message = fmt.Sprintf("%s \n", "No Neighborhoods found")
		rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
		return
	}
	//show all neighborhoods
	message = fmt.Sprintf("%s \n", "*You can find food trucks in following neighborhoods*")
	for _, n := range resp.Neighborhoods {
		message += fmt.Sprintf("• %s \n", n.ID)
	}
	rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
}

func showLocations(rtm *slack.RTM, text string, channel string) {
	var message string
	tokens := strings.Split(text, "in")
	if len(tokens) < 2 {
		rtm.SendMessage(rtm.NewOutgoingMessage("Missing Neighborhood", channel))
		return
	}
	neighborhood := strings.TrimSpace(tokens[1])
	if len(neighborhood) == 0 {
		rtm.SendMessage(rtm.NewOutgoingMessage("Missing neighborhood", channel))
		return
	}
	p, _ := seattlefoodtruck.NewProxy("https://www.seattlefoodtruck.com")
	lr := seattlefoodtruck.LocationRequest{
		Page:         1,
		Neighborhood: neighborhood,
	}
	resp, err := p.GetLocations(&lr)
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(err.Error(), channel))
		return
	}
	if len(resp.Locations) == 0 {
		message = fmt.Sprintf("No locations found at %s neighborhood \n", neighborhood)
		rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
		return
	}
	message = fmt.Sprintf("%s \n", "*You can find food trucks in following locations*")
	for _, l := range resp.Locations {
		message += fmt.Sprintf("• %s - %v \n", l.Name, l.UID)
	}
	rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
}

func showTrucks(rtm *slack.RTM, text string, channel string) {
	var message string
	//extract location id from text
	tokens := strings.Split(text, "at")
	if len(tokens) < 2 {
		rtm.SendMessage(rtm.NewOutgoingMessage("Missing location", channel))
		return
	}
	locString := strings.TrimSpace(tokens[1])
	if len(locString) == 0 {
		rtm.SendMessage(rtm.NewOutgoingMessage("Missing location", channel))
		return
	}
	location, _ := strconv.Atoi(locString)
	p, _ := seattlefoodtruck.NewProxy("https://www.seattlefoodtruck.com")
	req := seattlefoodtruck.NewLocationEventsRequest(location, 1)
	resp, err := p.GetLocationEvents(&req)
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(err.Error(), channel))
		return
	}
	if len(resp.Events) == 0 {
		message = fmt.Sprintf("No events at %v", locString)
		rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
		return
	}
	index := find(resp.Events, filterByStartDate)
	if index == -1 {
		message = fmt.Sprintf("No food trucks found at %v", locString)
		rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
		return
	}
	event := resp.Events[0]
	fmt.Printf("Found Event : %v", event)
	st, _ := time.Parse(time.RFC3339, event.StartTime)
	et, _ := time.Parse(time.RFC3339, event.EndTime)
	_, m, d := st.Date()
	message = fmt.Sprintf("%v %v %v - %v \n", m, d, st.Hour(), et.Hour())

	for _, b := range event.Bookings {
		message += fmt.Sprintf("*%v* (%s) %v \n", b.Truck.Name,
			strings.Join(b.Truck.FoodCategories, ","), s3Bucket+b.Truck.FeaturedPhoto)
	}
	//send message to channel
	rtm.SendMessage(rtm.NewOutgoingMessage(message, channel))
}

//returns the index of found event for today's date, if none returns -1
func find(events []seattlefoodtruck.Event, f func(seattlefoodtruck.Event) bool) int {
	for i, e := range events {
		if f(e) {
			return i
		}
	}
	return -1
}

// used to filter by event start date
func filterByStartDate(event seattlefoodtruck.Event) bool {
	ct := time.Now()
	y1, m1, d1 := ct.Date()
	//parse start time
	st, _ := time.Parse(time.RFC3339, event.StartTime)
	y2, m2, d2 := st.Date()
	if y1 == y2 && m1 == m2 && d1 == d2 {
		return true
	}
	return false
}
