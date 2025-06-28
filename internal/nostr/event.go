package nostr

import (
	"context"
	"github.com/nbd-wtf/go-nostr"
	"log/slog"
	"sync"
	"time"
)

func PublishEvent(ctx context.Context, event nostr.Event, relays []string) {
	var wg sync.WaitGroup
	slog.Info("Attempting to publish event", "event_id", event.ID, "relay_count", len(relays))

	for _, url := range relays {
		wg.Add(1)
		go func(relayURL string) {
			defer wg.Done()

			publishCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			relay, err := nostr.RelayConnect(publishCtx, relayURL)
			if err != nil {
				slog.Error("Error connecting to relay", "relay", relayURL, "error", err)
				return
			}
			defer relay.Close()

			if err = relay.Publish(publishCtx, event); err != nil {
				slog.Error("Error publishing event to relay", "event_id", event.ID, "relay", relayURL, "error", err)
				return
			}

			slog.Info("Successfully published event to relay", "event_id", event.ID, "relay", relayURL)
		}(url)
	}

	wg.Wait()
	slog.Info("Finished publishing attempts", "event_id", event.ID)
}
