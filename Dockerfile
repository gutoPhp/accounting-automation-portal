FROM node:18-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend
WORKDIR /app
COPY go.mod ./
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /sheep ./cmd/server

FROM alpine:3.21
RUN addgroup -S sheep && adduser -S sheep -G sheep
WORKDIR /app
COPY --from=backend /sheep /app/sheep
COPY --from=frontend /app/web/dist /app/web/dist
RUN mkdir /app/data && chown -R sheep:sheep /app
USER sheep
ENV ADDR=:8080 DATA_FILE=/app/data/sheep.json
EXPOSE 8080
CMD ["/app/sheep"]

