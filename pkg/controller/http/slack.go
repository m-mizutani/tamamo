package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/m-mizutani/goerr"
	slack_ctrl "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/slack-go/slack/slackevents"
)

func slackEventHandler(ctrl *slack_ctrl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if controller is nil
		if ctrl == nil {
			slog.ErrorContext(r.Context(), "slack controller is nil")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handleError(w, r, goerr.Wrap(err, "failed to read request body"))
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			handleError(w, r, goerr.Wrap(err, "failed to parse slack event").With("body", string(body)))
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			// Handle Slack challenge for URL verification
			var response *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &response)
			if err != nil {
				handleError(w, r, goerr.Wrap(err, "failed to unmarshal slack challenge response").With("body", string(body)))
				return
			}
			w.Header().Set("Content-Type", "text")
			if _, err := w.Write([]byte(response.Challenge)); err != nil {
				slog.ErrorContext(r.Context(), "failed to write challenge response", "error", err)
			}
			slog.InfoContext(r.Context(), "slack URL verification succeeded")

		case slackevents.CallbackEvent:
			// Handle actual Slack events
			innerEvent := eventsAPIEvent.InnerEvent

			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				if err := ctrl.HandleSlackAppMention(r.Context(), &eventsAPIEvent, ev); err != nil {
					slog.ErrorContext(r.Context(), "failed to handle app mention", "error", err)
					// Return 200 to prevent Slack retry
				}

			case *slackevents.MessageEvent:
				if err := ctrl.HandleSlackMessage(r.Context(), &eventsAPIEvent, ev); err != nil {
					slog.ErrorContext(r.Context(), "failed to handle message", "error", err)
					// Return 200 to prevent Slack retry
				}

			default:
				slog.WarnContext(r.Context(), "unknown event type", "event", ev, "body", string(body))
			}

			// Always return 200 for callback events
			w.WriteHeader(http.StatusOK)

		default:
			slog.WarnContext(r.Context(), "unknown slack event type", "type", eventsAPIEvent.Type)
			w.WriteHeader(http.StatusOK)
		}
	}
}
