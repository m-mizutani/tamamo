package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	slack_ctrl "github.com/m-mizutani/tamamo/pkg/controller/slack"
	"github.com/m-mizutani/tamamo/pkg/utils/async"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
	"github.com/slack-go/slack/slackevents"
)

func slackEventHandler(ctrl *slack_ctrl.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if controller is nil
		if ctrl == nil {
			err := goerr.New("slack controller is nil")
			errors.Handle(r.Context(), err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			err = goerr.Wrap(err, "failed to read request body")
			errors.Handle(r.Context(), err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			err = goerr.Wrap(err, "failed to parse slack event", goerr.V("body", string(body)))
			errors.Handle(r.Context(), err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			// Handle Slack challenge for URL verification
			var response *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &response)
			if err != nil {
				err = goerr.Wrap(err, "failed to unmarshal slack challenge response", goerr.V("body", string(body)))
				errors.Handle(r.Context(), err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "text")
			if _, err := w.Write([]byte(response.Challenge)); err != nil {
				errors.Handle(r.Context(), goerr.Wrap(err, "failed to write challenge response"))
			}
			ctxlog.From(r.Context()).Info("slack URL verification succeeded")

		case slackevents.CallbackEvent:
			// Handle actual Slack events asynchronously
			innerEvent := eventsAPIEvent.InnerEvent

			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				// Process app mention asynchronously
				eventsCopy := eventsAPIEvent
				evCopy := *ev
				async.Dispatch(r.Context(), func(ctx context.Context) error {
					return ctrl.HandleSlackAppMention(ctx, &eventsCopy, &evCopy)
				})

			case *slackevents.MessageEvent:
				// Process message asynchronously
				eventsCopy := eventsAPIEvent
				evCopy := *ev
				async.Dispatch(r.Context(), func(ctx context.Context) error {
					return ctrl.HandleSlackMessage(ctx, &eventsCopy, &evCopy)
				})

			default:
				ctxlog.From(r.Context()).Warn("unknown event type", "event", ev, "body", string(body))
			}

			// Immediately return 200 for callback events
			w.WriteHeader(http.StatusOK)

		default:
			ctxlog.From(r.Context()).Warn("unknown slack event type", "type", eventsAPIEvent.Type)
			w.WriteHeader(http.StatusOK)
		}
	}
}
