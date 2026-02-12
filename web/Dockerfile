FROM node AS build

WORKDIR /app

# Copy package.json and bun.lockb (if you're using bun.lockb instead of package-lock.json)
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application code
COPY . .

# Build the application
RUN npm run build

# Copy Nginx configuration
RUN mkdir -p /etc/nginx/conf.d/
COPY nginx.conf /etc/nginx/conf.d/default.conf

FROM nginxinc/nginx-unprivileged
WORKDIR /usr/share/nginx/html

# Copy the built files from the build stage
COPY --from=build --chown=101 /app/build /usr/share/nginx/html
COPY --from=build --chown=101 /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf

# Use a non-root user
USER 1000

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]