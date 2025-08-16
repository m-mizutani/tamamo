package slack

import (
	"context"

	"github.com/m-mizutani/ctxlog"
	"github.com/slack-go/slack/slackevents"
)

type Thread struct {
	TeamID    string `json:"team_id"`
	ChannelID string `json:"channel_id"`
	ThreadID  string `json:"thread_id"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Mention struct {
	UserID  string
	Message string
}

type Message struct {
	id       string
	channel  string
	threadID string
	teamID   string
	user     User
	msg      string
	ts       string
	mentions []Mention
}

func (x *Message) Thread() Thread {
	th := Thread{
		TeamID:    x.teamID,
		ChannelID: x.channel,
		ThreadID:  x.threadID,
	}
	if th.ThreadID == "" {
		th.ThreadID = x.id
	}
	return th
}

func (x *Message) ID() string {
	return x.id
}

func (x *Message) Mention() []Mention {
	return x.mentions
}

func (x *Message) User() *User {
	if x.user.ID == "" {
		return nil
	}
	return &x.user
}

func (x *Message) Text() string {
	return x.msg
}

func (x *Message) Timestamp() string {
	return x.ts
}

func (x *Message) ChannelID() string {
	return x.channel
}

func (x *Message) ThreadID() string {
	if x.threadID == "" {
		return x.id
	}
	return x.threadID
}

func (x *Message) TeamID() string {
	return x.teamID
}

func (x *Message) InThread() bool {
	return x.threadID != ""
}

func NewMessage(ctx context.Context, ev *slackevents.EventsAPIEvent) *Message {
	switch inEv := ev.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		return &Message{
			id:       inEv.TimeStamp,
			channel:  inEv.Channel,
			threadID: inEv.ThreadTimeStamp,
			teamID:   ev.TeamID,
			user: User{
				ID:   inEv.User,
				Name: inEv.User, // TODO: get user name from Slack API
			},
			msg:      inEv.Text,
			ts:       inEv.TimeStamp,
			mentions: ParseMention(inEv.Text),
		}

	case *slackevents.MessageEvent:
		return &Message{
			id:       inEv.TimeStamp,
			channel:  inEv.Channel,
			threadID: inEv.ThreadTimeStamp,
			teamID:   ev.TeamID,
			user: User{
				ID:   inEv.User,
				Name: inEv.User, // TODO: get user name from Slack API
			},
			msg:      inEv.Text,
			ts:       inEv.TimeStamp,
			mentions: ParseMention(inEv.Text),
		}

	default:
		ctxlog.From(ctx).Warn("unknown event type", "event", inEv)
		return nil
	}
}
