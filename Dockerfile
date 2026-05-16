# syntax=docker/dockerfile:1.7

# --- Stage 1: build the SPA ---
FROM node:22-alpine AS web-build
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ .
RUN npm run build

# --- Stage 2: build the Go binary with the SPA embedded ---
FROM golang:1.26-alpine AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Replace the placeholder dist with the freshly built SPA.
RUN rm -rf internal/static/dist
COPY --from=web-build /web/dist internal/static/dist
ARG VERSION=0.0.0-dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w \
        -X github.com/Exonical/kubevirt-management/internal/version.Version=${VERSION} \
        -X github.com/Exonical/kubevirt-management/internal/version.Commit=${COMMIT} \
        -X github.com/Exonical/kubevirt-management/internal/version.BuildDate=${BUILD_DATE}" \
    -o /out/kubevirt-management ./cmd/server

# --- Stage 3: minimal runtime image ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-build /out/kubevirt-management /usr/local/bin/kubevirt-management
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/kubevirt-management"]
