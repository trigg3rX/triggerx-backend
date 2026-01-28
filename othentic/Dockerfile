FROM node:22.6

RUN npm install -g npm@10.5.0

WORKDIR /app

RUN npm i -g @othentic/node@1.1.2

# Copy the delayed start script (build context is parent directory)
COPY scripts/docker/dev/delayed-start.sh /usr/local/bin/delayed-start.sh
RUN chmod +x /usr/local/bin/delayed-start.sh

ENTRYPOINT [ "otnode" ]
