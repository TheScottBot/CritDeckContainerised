FROM golang:1.26.2

ENV TOKEN ""

WORKDIR /app
COPY ./app/go.mod ./
COPY ./app/go.sum ./

RUN go mod download

COPY ./app/main.go ./
COPY ./app/playerDeck.json ./

RUN go build -o /app/critdeck

ENTRYPOINT ["/app/critdeck"]