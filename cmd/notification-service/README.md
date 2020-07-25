# Notification Service

## Subscription

Setup GraphQL subscription
```
subscription onSubs {
  pushNotif {
    helloSaid {
      id, msg
    }
  }
}
```

Broadcast message
```
mutation {
  pushNotif {
    sayHello(msg: "Hai") {
      id, msg
    }
  }
}
```

## Push Notification (using firebase)

### Via REST

```
POST: {{host}}/v1/pushnotif/push
Authorization: Basic {{key}}
Content-Type: application/json
Body:
{
    "to": "deviceID", 
    "title": "Hallo", 
    "message": "hai"
}
```

### Via GraphQL

Open GraphQL playground -> `{{host}}/graphql/playground`
```
mutation {
    pushNotif {
        push(payload: {
            to: "deviceID", 
            title: "Hallo", 
            message: "hai"
        })
    }
}
```


## Scheduled Push Notification

### Via REST

```
POST: {{host}}/v1/pushnotif/schedule
Authorization: Basic {{key}}
Content-Type: application/json
Body:
{
    "scheduledAt": "2020-06-26T00:00:00+07:00",
    "data" : {
        "to": "deviceID", 
        "title": "Hallo", 
        "message": "hai"
    }
}
```

### Via GraphQL

Open GraphQL playground -> `{{host}}/graphql/playground`
```
mutation {
    pushNotif {
        scheduledNotification(payload: {
            scheduledAt: "2020-06-26T00:00:00+07:00",
            data: {
                to: "deviceID", 
                title: "Hallo", 
                message: "hai"
            }
        })
    }
}
```
