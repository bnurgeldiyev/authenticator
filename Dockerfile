FROM alpine:latest
WORKDIR /web/teswir-go
COPY . .
EXPOSE 8081
CMD ["./daemon"]