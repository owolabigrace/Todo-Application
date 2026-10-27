FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /todo-api .

FROM alpine:3
COPY --from=build /todo-api /todo-api
EXPOSE 8080
CMD ["/todo-api"]