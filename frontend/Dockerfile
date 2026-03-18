FROM node:18-bullseye-slim AS build

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm install

COPY . .

# CRA embeds env vars at build-time. For container demo we rely on nginx /api proxy.
ENV REACT_APP_API_URL=/api

RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80

