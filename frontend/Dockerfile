# Stage 1: Dependencies
FROM node:24-alpine AS dependencies
WORKDIR /app

COPY pnpm-lock.yaml package.json ./

RUN npm install -g pnpm && \
  pnpm install --frozen-lockfile --prod

# Stage 2: Builder
FROM node:24-alpine AS builder
WORKDIR /app

COPY pnpm-lock.yaml package.json ./

RUN npm install -g pnpm && \
  pnpm install --frozen-lockfile

COPY src ./src
COPY static ./static
COPY svelte.config.js vite.config.ts tsconfig.json ./

RUN pnpm run build

# Stage 3: Runtime
FROM node:24-alpine
WORKDIR /app

ENV NODE_ENV=production
ENV BASE_API=http://localhost:8080

# adapter-node otherwise trusts nothing and reads event.getClientAddress()
# straight off the TCP socket -- which, behind this platform's internal
# proxy, is that proxy's own address, not the visitor's. ADDRESS_HEADER
# tells it to read X-Forwarded-For instead, and XFF_DEPTH=1 says there is
# exactly one trusted hop (this platform's proxy) prepending to it, so the
# rightmost address in the header is trusted as the real client IP.
ENV ADDRESS_HEADER=X-Forwarded-For
ENV XFF_DEPTH=1

# Copy production dependencies from dependencies stage
COPY --from=dependencies /app/node_modules ./node_modules

# Copy built application
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./

EXPOSE 3000

CMD ["node", "build/index.js"]
