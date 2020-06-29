FROM node:14-alpine as build

WORKDIR /app
COPY package*.json yarn.lock ./
RUN yarn install
COPY tsconfig.json .
COPY public public
COPY src src
RUN yarn build

FROM nginx:1.19-alpine
ENV DCTNA_API https://host.docker.internal:8443
COPY nginx/templates /etc/nginx/templates/
COPY --from=build /app/build /usr/share/nginx/html
