package slack_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

func TestChannelType_EnumValues(t *testing.T) {
	// Test all valid channel type enum values
	validChannelTypes := []slack.ChannelType{
		slack.ChannelTypePublic,
		slack.ChannelTypePrivate,
		slack.ChannelTypeIM,
		slack.ChannelTypeMPIM,
	}

	for _, channelType := range validChannelTypes {
		gt.NotEqual(t, string(channelType), "")
	}

	// Test specific values
	gt.Equal(t, string(slack.ChannelTypePublic), "public")
	gt.Equal(t, string(slack.ChannelTypePrivate), "private")
	gt.Equal(t, string(slack.ChannelTypeIM), "im")
	gt.Equal(t, string(slack.ChannelTypeMPIM), "mpim")
}

func TestMessageType_EnumValues(t *testing.T) {
	// Test all valid message type enum values
	validMessageTypes := []slack.MessageType{
		slack.MessageTypeUser,
		slack.MessageTypeBot,
		slack.MessageTypeSystem,
	}

	for _, messageType := range validMessageTypes {
		gt.NotEqual(t, string(messageType), "")
	}

	// Test specific values
	gt.Equal(t, string(slack.MessageTypeUser), "user")
	gt.Equal(t, string(slack.MessageTypeBot), "bot")
	gt.Equal(t, string(slack.MessageTypeSystem), "system")
}

func TestDetermineChannelType(t *testing.T) {
	// Test public channel
	channelType := slack.DetermineChannelType(true, false, false, false)
	gt.Equal(t, channelType, slack.ChannelTypePublic)

	// Test private channel (group)
	channelType = slack.DetermineChannelType(false, true, false, false)
	gt.Equal(t, channelType, slack.ChannelTypePrivate)

	// Test IM channel
	channelType = slack.DetermineChannelType(false, false, true, false)
	gt.Equal(t, channelType, slack.ChannelTypeIM)

	// Test MPIM channel
	channelType = slack.DetermineChannelType(false, false, false, true)
	gt.Equal(t, channelType, slack.ChannelTypeMPIM)

	// Test unknown/default case (no flags set)
	channelType = slack.DetermineChannelType(false, false, false, false)
	gt.Equal(t, channelType, slack.ChannelTypePublic) // Should default to public
}

func TestDetermineMessageType(t *testing.T) {
	// Test user message
	messageType := slack.DetermineMessageType("U123456789", "")
	gt.Equal(t, messageType, slack.MessageTypeUser)

	// Test bot message
	messageType = slack.DetermineMessageType("", "B123456789")
	gt.Equal(t, messageType, slack.MessageTypeBot)

	// Test system message (no user ID or bot ID)
	messageType = slack.DetermineMessageType("", "")
	gt.Equal(t, messageType, slack.MessageTypeSystem)

	// Test edge case: both user and bot ID (should prioritize bot)
	messageType = slack.DetermineMessageType("U123456789", "B123456789")
	gt.Equal(t, messageType, slack.MessageTypeBot)
}

func TestChannelInfo_Structure(t *testing.T) {
	channelInfo := &slack.ChannelInfo{
		ID:        "C123456789",
		Name:      "general",
		Type:      slack.ChannelTypePublic,
		IsPrivate: false,
		UpdatedAt: time.Now(),
	}

	gt.Equal(t, channelInfo.ID, "C123456789")
	gt.Equal(t, channelInfo.Name, "general")
	gt.Equal(t, channelInfo.Type, slack.ChannelTypePublic)
	gt.Equal(t, channelInfo.IsPrivate, false)
	gt.NotEqual(t, channelInfo.UpdatedAt, time.Time{})
}

func TestAttachment_Structure(t *testing.T) {
	attachment := slack.Attachment{
		ID:       "file_123456789",
		Name:     "document.pdf",
		Mimetype: "application/pdf",
		FileType: "pdf",
		URL:      "https://files.slack.com/files/document.pdf",
	}

	gt.Equal(t, attachment.ID, "file_123456789")
	gt.Equal(t, attachment.Name, "document.pdf")
	gt.Equal(t, attachment.Mimetype, "application/pdf")
	gt.Equal(t, attachment.FileType, "pdf")
	gt.Equal(t, attachment.URL, "https://files.slack.com/files/document.pdf")
}

func TestSlackMessageLog_JSONMarshaling(t *testing.T) {
	originalTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)

	messageLog := &slack.SlackMessageLog{
		ID:          types.MessageID("msg_123456789"),
		TeamID:      "T123456789",
		ChannelID:   "C123456789",
		ChannelName: "general",
		ChannelType: slack.ChannelTypePublic,
		UserID:      "U123456789",
		UserName:    "testuser",
		BotID:       "",
		MessageType: slack.MessageTypeUser,
		Text:        "Hello, world!",
		Timestamp:   "1234567890.123456",
		ThreadTS:    "",
		Attachments: []slack.Attachment{
			{
				ID:       "file_123",
				Name:     "test.txt",
				Mimetype: "text/plain",
				FileType: "txt",
				URL:      "https://example.com/test.txt",
			},
		},
		CreatedAt: originalTime,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(messageLog)
	gt.NoError(t, err)
	gt.NotNil(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledLog slack.SlackMessageLog
	err = json.Unmarshal(jsonData, &unmarshaledLog)
	gt.NoError(t, err)

	// Verify unmarshaled data
	gt.Equal(t, unmarshaledLog.ID, messageLog.ID)
	gt.Equal(t, unmarshaledLog.ChannelID, messageLog.ChannelID)
	gt.Equal(t, unmarshaledLog.ChannelType, messageLog.ChannelType)
	gt.Equal(t, unmarshaledLog.MessageType, messageLog.MessageType)
	gt.Equal(t, unmarshaledLog.Text, messageLog.Text)
	gt.Equal(t, len(unmarshaledLog.Attachments), 1)
	gt.Equal(t, unmarshaledLog.Attachments[0].ID, "file_123")

	// Time should be preserved (within reasonable precision)
	timeDiff := unmarshaledLog.CreatedAt.Sub(originalTime)
	gt.True(t, timeDiff < time.Second && timeDiff > -time.Second)
}

func TestChannelInfo_JSONMarshaling(t *testing.T) {
	originalTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)

	channelInfo := &slack.ChannelInfo{
		ID:        "C123456789",
		Name:      "general",
		Type:      slack.ChannelTypePrivate,
		IsPrivate: true,
		UpdatedAt: originalTime,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(channelInfo)
	gt.NoError(t, err)
	gt.NotNil(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaledInfo slack.ChannelInfo
	err = json.Unmarshal(jsonData, &unmarshaledInfo)
	gt.NoError(t, err)

	// Verify unmarshaled data
	gt.Equal(t, unmarshaledInfo.ID, channelInfo.ID)
	gt.Equal(t, unmarshaledInfo.Name, channelInfo.Name)
	gt.Equal(t, unmarshaledInfo.Type, channelInfo.Type)
	gt.Equal(t, unmarshaledInfo.IsPrivate, channelInfo.IsPrivate)

	// Time should be preserved (within reasonable precision)
	timeDiff := unmarshaledInfo.UpdatedAt.Sub(originalTime)
	gt.True(t, timeDiff < time.Second && timeDiff > -time.Second)
}

func TestEnumValues_JSONMarshaling(t *testing.T) {
	// Test ChannelType JSON marshaling
	channelTypes := map[slack.ChannelType]string{
		slack.ChannelTypePublic:  "\"public\"",
		slack.ChannelTypePrivate: "\"private\"",
		slack.ChannelTypeIM:      "\"im\"",
		slack.ChannelTypeMPIM:    "\"mpim\"",
	}

	for channelType, expectedJSON := range channelTypes {
		jsonData, err := json.Marshal(channelType)
		gt.NoError(t, err)
		gt.Equal(t, string(jsonData), expectedJSON)

		// Test unmarshaling
		var unmarshaledType slack.ChannelType
		err = json.Unmarshal(jsonData, &unmarshaledType)
		gt.NoError(t, err)
		gt.Equal(t, unmarshaledType, channelType)
	}

	// Test MessageType JSON marshaling
	messageTypes := map[slack.MessageType]string{
		slack.MessageTypeUser:   "\"user\"",
		slack.MessageTypeBot:    "\"bot\"",
		slack.MessageTypeSystem: "\"system\"",
	}

	for messageType, expectedJSON := range messageTypes {
		jsonData, err := json.Marshal(messageType)
		gt.NoError(t, err)
		gt.Equal(t, string(jsonData), expectedJSON)

		// Test unmarshaling
		var unmarshaledType slack.MessageType
		err = json.Unmarshal(jsonData, &unmarshaledType)
		gt.NoError(t, err)
		gt.Equal(t, unmarshaledType, messageType)
	}
}

func TestSlackMessageLog_ComplexScenarios(t *testing.T) {
	t.Run("ThreadMessage", func(t *testing.T) {
		messageLog := &slack.SlackMessageLog{
			ID:          types.MessageID("msg_thread_123"),
			TeamID:      "T123456789",
			ChannelID:   "C123456789",
			ChannelName: "general",
			ChannelType: slack.ChannelTypePublic,
			UserID:      "U123456789",
			UserName:    "testuser",
			MessageType: slack.MessageTypeUser,
			Text:        "This is a thread reply",
			Timestamp:   "1234567890.123456",
			ThreadTS:    "1234567890.000000", // Parent thread timestamp
			CreatedAt:   time.Now(),
		}

		gt.NotEqual(t, messageLog.ThreadTS, "")
		gt.NotEqual(t, messageLog.ThreadTS, messageLog.Timestamp)
	})

	t.Run("BotMessageWithAttachment", func(t *testing.T) {
		messageLog := &slack.SlackMessageLog{
			ID:          types.MessageID("msg_bot_123"),
			TeamID:      "T123456789",
			ChannelID:   "C123456789",
			ChannelName: "random",
			ChannelType: slack.ChannelTypePublic,
			BotID:       "B123456789",
			MessageType: slack.MessageTypeBot,
			Text:        "Here's a file for you",
			Timestamp:   "1234567890.123456",
			Attachments: []slack.Attachment{
				{
					ID:       "file_bot_123",
					Name:     "report.pdf",
					Mimetype: "application/pdf",
					FileType: "pdf",
					URL:      "https://files.slack.com/files/report.pdf",
				},
			},
			CreatedAt: time.Now(),
		}

		gt.Equal(t, messageLog.UserID, "") // No user ID for bot messages
		gt.NotEqual(t, messageLog.BotID, "")
		gt.Equal(t, len(messageLog.Attachments), 1)
	})

	t.Run("PrivateChannelMessage", func(t *testing.T) {
		messageLog := &slack.SlackMessageLog{
			ID:          types.MessageID("msg_private_123"),
			TeamID:      "T123456789",
			ChannelID:   "G123456789", // Private group ID
			ChannelName: "secret-project",
			ChannelType: slack.ChannelTypePrivate,
			UserID:      "U123456789",
			UserName:    "testuser",
			MessageType: slack.MessageTypeUser,
			Text:        "Confidential information",
			Timestamp:   "1234567890.123456",
			CreatedAt:   time.Now(),
		}

		gt.Equal(t, messageLog.ChannelType, slack.ChannelTypePrivate)
		gt.True(t, messageLog.ChannelID[0] == 'G') // Private groups start with G
	})

	t.Run("DirectMessage", func(t *testing.T) {
		messageLog := &slack.SlackMessageLog{
			ID:          types.MessageID("msg_dm_123"),
			TeamID:      "T123456789",
			ChannelID:   "D123456789", // DM channel ID
			ChannelName: "@testuser",
			ChannelType: slack.ChannelTypeIM,
			UserID:      "U123456789",
			UserName:    "testuser",
			MessageType: slack.MessageTypeUser,
			Text:        "Direct message",
			Timestamp:   "1234567890.123456",
			CreatedAt:   time.Now(),
		}

		gt.Equal(t, messageLog.ChannelType, slack.ChannelTypeIM)
		gt.True(t, messageLog.ChannelID[0] == 'D') // DM channels start with D
	})
}
