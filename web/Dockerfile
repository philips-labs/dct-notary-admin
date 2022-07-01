FROM node:16-alpine as build

WORKDIR /app
COPY package*.json yarn.lock ./
RUN yarn install
COPY tsconfig.json craco.config.js tailwind.config.js tsconfig.json ./
COPY public public
COPY src src
RUN yarn build

FROM nginx:1.23.0-alpine
ENV DCTNA_API https://host.docker.internal:8443
COPY nginx/templates /etc/nginx/templates/
COPY --from=build /app/build /usr/share/nginx/html
