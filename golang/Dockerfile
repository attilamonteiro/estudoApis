FROM golang:1.23

WORKDIR /golang

COPY . .

RUN go get -d -v ./...

RUN go build -o golang .

EXPOSE 8000

CMD ["./golang"]