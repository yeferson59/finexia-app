# Stage 1: Builder
FROM ghcr.io/pnpm/pnpm:11 AS builder
RUN pnpm runtime set node 24 -g

WORKDIR /app

COPY pnpm-lock.yaml package.json ./

RUN pnpm install --frozen-lockfile --prod

COPY src ./src
COPY static ./static
COPY svelte.config.js vite.config.ts tsconfig.json ./

RUN pnpm run build

# Stage 3: Runtime
FROM node:24-alpine
WORKDIR /app

ENV NODE_ENV=production
ENV BASE_API=http://localhost:8080
ENV ADDRESS_HEADER=X-Forwarded-For
ENV XFF_DEPTH=1

# Copy production dependencies from builder stage
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./

EXPOSE 3000

CMD ["node", "build/index.js"]
