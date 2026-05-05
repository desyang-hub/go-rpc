FROM node:20-alpine

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json* ./
RUN npm install 2>/dev/null || npm install --package-lock-only 2>/dev/null || npm install

# Copy application code
COPY . .

# Run the client
CMD ["npx", "ts-node", "client.ts"]
