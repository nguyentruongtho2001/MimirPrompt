FROM alpine:latest

WORKDIR /pb

# Download PocketBase (compatible with both architectures)
ARG PB_VERSION=0.23.4
ARG TARGETARCH

RUN apk add --no-cache wget unzip ca-certificates \
    && wget -q https://github.com/pocketbase/pocketbase/releases/download/v${PB_VERSION}/pocketbase_${PB_VERSION}_linux_${TARGETARCH}.zip \
    && unzip pocketbase_${PB_VERSION}_linux_${TARGETARCH}.zip \
    && rm pocketbase_${PB_VERSION}_linux_${TARGETARCH}.zip \
    && chmod +x pocketbase

# Copy hooks if any
COPY pb_hooks ./pb_hooks

EXPOSE 8090

CMD ["./pocketbase", "serve", "--http=0.0.0.0:8090"]
