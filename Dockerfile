FROM golang:1.19 AS build-stage
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/wserver cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/wclient cmd/client/main.go

FROM scratch
COPY --from=build-stage /app/wserver /bin/wserver
COPY --from=build-stage /app/wclient /bin/wclient
EXPOSE 9000
CMD ["/bin/wserver"]
