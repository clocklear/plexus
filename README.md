# plexus
An experimental project leveraging Plex media server webhooks in order to facilitate other integrations

The goal here for me is to create a webapp that will allow me to land Plex webhook calls, explore them in a simple UI, and perhaps rig up automation based on them. (perhaps dim my lights when a movie starts on a certain device?)

## current state

Currently, Plexus can fire its own webhooks based on webhook input from Plex.  You'll need a `config.json` that looks something like this:

```
{
  "triggers": [
    {
      "properties": {
        "event": "media.play",
        "Player.uuid": "change.me"
      },
      "actions": [
        {
          "type": "webhook",
          "config": {
            "action": "POST",
            "url": "https://some.url.that.should.do.something"
          }
        }
      ]
    },
    {
      "properties": {
        "event": "media.stop",
        "Player.uuid": "change.me"
      },
      "actions": [
        {
          "type": "webhook",
          "config": {
            "action": "POST",
            "url": "https://another.url.that.changes.the.world"
          }
        }
      ]
    }
  ]
}
```

Triggers should be a list of things Plexus should respond to.  Each trigger has a `properties` node that will be matched against the activity coming out of Plex.  If all properties match, the trigger is considered a match, and actions are evaluated.  Note that the keys of `properties` can be deep references to complex objects in the payload body.  Use period to indicated nesting.

Each trigger has a corresponding list of `actions` that will be fired if the trigger is considered a match.  Currently the only supported action is `webhook`, and it is very simple -- you can only control the URL and the HTTP verb used in the request.  Still, this is very powerful.

As a proof of concept, I have been able to use Plexus to monitor activity from my Plex server and on media plays/stops originating from my living room player, I can dim the living room lights accordingly.  This is accomplished by invoking IFTTT webhooks that can talk to my Wemo devices remotely.

YMMV.  Very WIP.
