FROM node:20.18.2-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM caddy:2.9.1-alpine
COPY --from=build /src/dist /usr/share/caddy
