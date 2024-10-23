FROM node:20

WORKDIR /keeperApp

COPY package.json ./
COPY utils ./utils
COPY keeper ./keeper
COPY ./keeper/.env.example ./.env

RUN yarn install

EXPOSE 3000

CMD ["node", "keeper/main.js", "$KEEPER_ID"]