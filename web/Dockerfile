FROM node:24-alpine AS build

WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY index.html tsconfig.json vite.config.ts ./
COPY public public
COPY src src
RUN pnpm build

FROM nginx:1.31.2-alpine
ENV DCTNA_API https://host.docker.internal:8443
COPY nginx/templates /etc/nginx/templates/
COPY --from=build /app/dist /usr/share/nginx/html
