FROM golang:1.13.4-alpine

WORKDIR /app

COPY fluffy /app/fluffy
COPY home.html /app/home.html

EXPOSE 8080

CMD ["./fluffy_chat"]
