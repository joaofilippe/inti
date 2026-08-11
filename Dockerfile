FROM golang:alpine AS builder

ARG GITHUB_TOKEN

RUN apk add --no-cache git

WORKDIR /app
RUN git clone https://${GITHUB_TOKEN}@github.com/joaofilippe/inti.git .
RUN go mod tidy
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o inti-app main.go

FROM alpine:latest
WORKDIR /app

# Copia o binário
COPY --from=builder /app/inti-app .
# Copia os templates DOCX que a aplicação precisa
COPY --from=builder /app/*.docx ./

EXPOSE 8080

CMD ["./inti-app"]
