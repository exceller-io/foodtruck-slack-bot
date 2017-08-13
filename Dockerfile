FROM golang as builder

WORKDIR /go/src/github.com/rprakashg/foodtruck-slack-bot/

RUN go get -d -v github.com/nlopes/slack \
    && go get -d -v github.com/robfig/cron

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o foodtruck-slack-bot .

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/rprakashg/foodtruck-slack-bot/foodtruck-slack-bot .

CMD [ "./foodtruck-slack-bot" ]