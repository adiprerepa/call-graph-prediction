FROM golang:1.20-alpine
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN CGO_ENABLED=0 GOOS=linux go build -o main .
CMD ["/app/main"]

EXPOSE 8080