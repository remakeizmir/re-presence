# The reporting endpoint

One request, repeated. Any editor, any language, twenty lines of code.

```
POST https://api.remakeizmir.com/api/v1/presence/editor
Authorization: Bearer rmk_dev_…
Content-Type: application/json

{
  "app": "Zed",
  "project": "server",
  "file": "chat_service.go",
  "line": 412,
  "language": "go",
  "debugging": false,
  "idle": false
}
```

Every field but `app` is optional. A report with nothing else still says "this
person is working".

- **Send one every 30 seconds** while the editor has focus. The hub forgets a
  report after two minutes, so anything slower flickers.
- **`"idle": true`** takes the card down at once — send it when the editor loses
  focus for a while, or on exit.
- **Paths are cut to their last segment** by the server, whatever you send. Send
  the short form anyway; there is no reason for a path to leave the machine.
- The clock on the card keeps running across files. A gap of more than ten
  minutes starts it again.
- Rate limit: 240 requests a minute per address.

Answers: `200` on success, `401` when the key is wrong or revoked.
