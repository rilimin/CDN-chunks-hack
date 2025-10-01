# CDN chunks hack
Totally legit

# How to use

1. Change infos inside config.toml
2. Run `go run .` to start server in localhost

# config.toml explain
1. Port: of the server

2. ChunkSize: chunk size in bytes

3. MaxRoutine: maximum goroutine the server may create and use

4. Key: XOR key to prevent normal file viewing

5. WebhookPath: path to text file contains webhooks. Webhooks are sperated by line

- For example:

```
https://discord.com/api/webhooks/...
https://discord.com/api/webhooks/...
https://discord.com/api/webhooks/...
...
```

6. UserToken: Discord user token to refresh CDN 

# Credit
Video handler: https://www.zeng.dev/post/2023-http-range-and-play-mp4-in-browser/

Semaphore: https://github.com/jamesrr39/semaphore/blob/master/semaphore.go

Refresh CDN link: https://github.com/ShufflePerson/Discord_CDN
