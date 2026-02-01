# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Package files kopyala
COPY package*.json ./
RUN rm -f package-lock.json
RUN npm install

# Kaynak kodları kopyala
COPY . .

# Build yap
RUN npm run build

# Runtime stage
FROM node:18-alpine

WORKDIR /app

# Build output'u kopyala
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package*.json ./

# Production dependencies
RUN npm install --production

# Port expose et
EXPOSE 3000

# Environment variable
ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

# Uygulamayı çalıştır
CMD ["node", ".output/server/index.mjs"]
