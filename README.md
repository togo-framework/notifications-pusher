# notifications-pusher

> **Pusher Channels** broadcast channel for [togo](https://to-go.dev) **notifications**.

Delivers a `BroadcastNotification` to Pusher's hosted realtime service (signed REST API).

## Install

```bash
togo install togo-framework/notifications-pusher
```

```ini
PUSHER_APP_ID=...
PUSHER_KEY=...
PUSHER_SECRET=...
PUSHER_CLUSTER=eu
```

Registers the `pusher` notifications channel. Use an event of the form `channel:name` to target a Pusher channel. MIT
