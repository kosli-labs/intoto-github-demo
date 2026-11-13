FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /hello .

FROM scratch
COPY --from=build /hello /hello
USER 65534
EXPOSE 8080
ENTRYPOINT ["/hello"]
