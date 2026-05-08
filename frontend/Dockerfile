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

ENV BASE_API=http://localhost:8080

RUN pnpm run build

# Stage 3: Runtime
FROM node:24-alpine
WORKDIR /app

ENV NODE_ENV=production

# Copy production dependencies from dependencies stage
COPY --from=dependencies /app/node_modules ./node_modules

# Copy built application
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./

EXPOSE 3000

CMD ["node", "build/index.js"]
