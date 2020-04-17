# Line Chatbot Service

GraphQL query for find all event log:
```sh
query {
    event {
        getAll(filter: {page:1, limit: 10, sort: "desc", orderBy: "timestamp"}) {
            meta { page, limit, totalRecords, totalPages }
            data { 
                id, replyToken, type, timestamp, sourceId, sourceType, error, 
                message { id, type, text, response } 
            }
        }
    }
}
```

## Click this link for add my line bot channel
<a href="https://line.me/R/ti/p/%40ylf0312k"><img height="36" border="0" alt="Tambah Teman" src="https://scdn.line-apps.com/n/line_add_friends/btn/en.png"></a>

See https://github.com/agungdwiprasetyo/chatbot-ai (repository text mining for processing input chat message)