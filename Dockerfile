# --- build stage ---
    FROM golang:1.22-alpine AS build

    WORKDIR /src
    
    # git + ca-certificates 
    RUN apk add --no-cache git ca-certificates
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    # static-ish binary
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/scanner-api ./cmd/api
    
    # --- runtime stage ---
    FROM gcr.io/distroless/static:nonroot
    
    WORKDIR /app
    COPY --from=build /out/scanner-api /app/scanner-api
    
    EXPOSE 8080
    
    # PORT, default 8080
    ENTRYPOINT ["/app/scanner-api"]
    