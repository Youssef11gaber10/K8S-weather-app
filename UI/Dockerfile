FROM node:17
WORKDIR /usr/src/app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
ENTRYPOINT ["node"]
CMD [ "app.js" ]
# will enhance this later
# its not react app its node app with some static files so can't serve it on nginx it need node server run time 