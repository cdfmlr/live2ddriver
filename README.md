# live2ddriver

> 🔑 This is a muli component. You can find the main repository [here](https://github.com/cdfmlr/muvtuber).

这个程序驱动 [live2dview](https://github.com/cdfmlr/live2dview) 的运行.

Goto the [main repository](https://github.com/cdfmlr/muvtuber) for more information.

## Usage

Live2ddriver is designed to be a simple message forwarder:

- receive requests from end-user or higher-level driver via http
- forward requests to live2dview via websocket
   - and then live2dview will behave according to the requests

```
USER --http--> LIVE2DDRIVER --websocket--> LIVE2DVIEW
```

In the early stage of development, live2ddriver is also responsible for analyzing the emotion of the text, mapping the emotion to the expression / motion of the specific live2d model, and sending the result expression / motion to live2dview. The emotion analysis is dependent on the [emotext](https://github.com/murchinroom/emotext).

> It was designed to write individual drivers for each live2d models. But it's not a good idea. I'm trying to make it more general, that is, make it possible to *configure* the emotion mapping for each live2d model, so that we can use a single driver for all live2d models. But it's not done yet.

### Run

```sh
go run . -wsAddr 0.0.0.0:9001 -httpAddr 0.0.0.0:9002  -shizuku 0.0.0.0:9004 -verbose
```

### Docker Compose

```yaml
  live2ddriver:
    image: murchinroom/live2ddriver:v0.1.0-alpha.1
    build: ./live2ddriver/
    ports:
      - "51071:9001"
      - "51072:9002"
      - "51074:9004"
    environment:
      - EMOTEXT_SERVER=http://emotext:9003
    depends_on:
      - emotext
```

### Ports

- `51071`: `9001`: listen & serve websocket: live2dview connect to this port to get requests from the driver.
- `51072`: `9002`: listen & serve http: end-user or higher-level driver connect to this port to send requests to the driver. (requests are forwarded to live2dview via websocket (9001))
- `51074`: `9004`: listen & serve shizuku: end-user or higher-level driver connect to this port to send text to the shizuku driver (a specific driver for the example shizuku live2d model). Live2ddriver will analyze the emotion of the text and send the result expression / motion to live2dview via websocket (9001).

### Live2dRequests

Things that you can send to 9002, and live2ddriver will forward to live2dview:

- `{"model": "url to the an live2d model src"}`
- `{"motion": "motion group name"}`
- `{"expression": "expression id (name or index)"}`
- `{"speak": { "audio":  "audio src", "expression": "expression id", "motion": "motion group" }}`
   - `audio src` can be an url to audio file (wav or mp3) or a base64 encoded data (data:audio/wav;base64,xxxx)

## License

live2ddriver is licensed under the MIT license.

