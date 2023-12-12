FROM ncr.sky.nhn.no/dockerhub/node:alpine AS build

WORKDIR /app
COPY package.json package-lock.json .
RUN npm install --force
COPY . .

RUN npm run build
RUN mkdir -p /etc/nginx/conf.d/ && echo "\
server {\
    listen 8080;\
    server_name localhost;\
    root /usr/share/nginx/html;\
    index index.html;\
\
    location / {\
        try_files \$uri \$uri/ /index.html;\
    }\
\
    error_page 500 502 503 504 /50x.html;\
    location = /50x.html {\
        root /usr/share/nginx/html;\
    }\
}\
" > /etc/nginx/conf.d/default.conf

FROM ncr.sky.nhn.no/dockerhub/nginxinc/nginx-unprivileged
WORKDIR /usr/share/nginx/html

# Copy the built files from the build stage
COPY --from=build --chown=101 /app/build /usr/share/nginx/html
COPY --from=build --chown=101 /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf

# Use a non-root user
USER 1000

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]