FROM node:26-alpine AS base

WORKDIR /app

# Install yarn
RUN apk add --no-cache yarn

# Copy package files
COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn ./.yarn

# Install dependencies
RUN yarn install --immutable

# Copy source code
COPY . .

# Build TypeScript
RUN yarn build

# Prune to production dependencies
FROM base AS pruned
RUN yarn workspaces focus --production

# Production stage
FROM node:26-alpine AS production

ENV NODE_ENV=production

WORKDIR /app

# Copy package files
COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn ./.yarn

# Copy pruned node_modules from pruned stage
COPY --from=pruned /app/node_modules ./node_modules

# Copy built application
COPY --from=base /app/dist ./dist

EXPOSE 3001

# Health check - checks every 30s with 3s timeout, 3 retries before unhealthy
COPY healthcheck.cjs ./
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node /app/healthcheck.cjs

# Create non-root user and set ownership of /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

CMD ["node", "dist/index.mjs"]