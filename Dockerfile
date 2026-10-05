FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/gosysmon ./cmd/gosysmon

FROM scratch
COPY --from=build /out/gosysmon /gosysmon
EXPOSE 9090
ENTRYPOINT ["/gosysmon"]
CMD ["--serve", "--addr", ":9090"]
