FROM node:22.6

RUN npm install -g npm@10.5.0

WORKDIR /app

RUN npm i -g @othentic/othentic-cli@1.13.0

ENTRYPOINT [ "othentic-cli" ]
