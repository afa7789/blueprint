FROM oven/bun:latest

WORKDIR /app

COPY package.json bun.lock* ./
RUN bun install

COPY . .

EXPOSE 5173

# Install deps (in case package.json changed via volume mount) then start dev server
CMD ["sh", "-c", "bun install && bun run dev --host 0.0.0.0"]
