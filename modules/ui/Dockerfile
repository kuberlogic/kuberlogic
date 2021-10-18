# Stage 1: Compile and Build angular codebase

# Use official node image as the base image
FROM node:16.3.0 as build

# Set the working directory
WORKDIR /usr/local/app

# Copy dependency information
COPY ./package.json ./package-lock.json ./

# Install all the dependencies
RUN npm install --loglevel verbose

# Copy app source code
COPY . ./

# Generate the build of the application
RUN npm run build:production

# Stage 2: Serve app with nginx server

# Use official nginx image as the base image
FROM nginx:latest

# Copy the build output to replace the default nginx contents.
COPY --from=build /usr/local/app/dist/kuberlogic /usr/share/nginx/html
COPY docker/ngx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80
