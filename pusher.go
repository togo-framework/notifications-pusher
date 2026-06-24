// Package pusher is a Pusher Channels broadcast channel for togo notifications.
// A BroadcastNotification is delivered to Pusher's hosted realtime service.
// Install: `togo install togo-framework/notifications-pusher`.
// Env: PUSHER_APP_ID, PUSHER_KEY, PUSHER_SECRET, PUSHER_CLUSTER.
package pusher

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/togo-framework/notifications"
	"github.com/togo-framework/togo"
)

func init() {
	notifications.RegisterChannel("pusher", func(k *togo.Kernel) notifications.Channel {
		return &channel{
			appID:   os.Getenv("PUSHER_APP_ID"),
			key:     os.Getenv("PUSHER_KEY"),
			secret:  os.Getenv("PUSHER_SECRET"),
			cluster: os.Getenv("PUSHER_CLUSTER"),
			client:  &http.Client{Timeout: 15 * time.Second},
		}
	})
}

type channel struct {
	appID, key, secret, cluster string
	client                      *http.Client
}

func (c *channel) Send(ctx context.Context, to notifications.Notifiable, n notifications.Notification) error {
	bn, ok := n.(notifications.BroadcastNotification)
	if !ok {
		return nil
	}
	if c.appID == "" || c.key == "" || c.secret == "" {
		return nil
	}
	event, data := bn.ToBroadcast(to)
	ch, name := "togo", event
	if i := strings.IndexByte(event, ':'); i >= 0 {
		ch, name = event[:i], event[i+1:]
	}
	dataJSON, _ := json.Marshal(data)
	body, _ := json.Marshal(map[string]any{"name": name, "channel": ch, "data": string(dataJSON)})

	cluster := c.cluster
	if cluster == "" {
		cluster = "mt1"
	}
	path := "/apps/" + c.appID + "/events"
	params := url.Values{}
	params.Set("auth_key", c.key)
	params.Set("auth_timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("auth_version", "1.0")
	params.Set("body_md5", fmt.Sprintf("%x", md5.Sum(body)))
	toSign := "POST\n" + path + "\n" + params.Encode()
	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write([]byte(toSign))
	params.Set("auth_signature", hex.EncodeToString(mac.Sum(nil)))

	endpoint := "https://api-" + cluster + ".pusher.com" + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notifications-pusher: status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
