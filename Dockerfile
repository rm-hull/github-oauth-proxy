FROM node:22-alpine AS base

# Enable Corepack for Yarn Berry
RUN corepack enable

WORKDIR /app

# Copy package files
COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn ./.yarn

# Install dependencies
RUN yarn install --immutable

# Copy source code
COPY . .

# Build TypeScript
RUN yarn build

# Production stage
FROM node:22-alpine AS production

RUN corepack enable

WORKDIR /app

# Copy package files
COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn ./.yarn

# Install production dependencies only
RUN yarn workspaces focus --production

# Copy built application
COPY --from=base /app/dist ./dist

EXPOSE 3001

# Health check - checks every 30s with 3s timeout, 3 retries before unhealthy
COPY healthcheck.js ./
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node /app/healthcheck.js

# Create non-root user and set ownership of /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

CMD ["node", "dist/index.js"]