FROM golang:1.25.5-alpine

WORKDIR /app

COPY . .

CMD ["go", "run", "."]
