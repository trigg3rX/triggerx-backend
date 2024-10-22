FROM node:20

WORKDIR /keeperApp

COPY package.json ./

RUN yarn install

COPY keeper /keeperApp

EXPOSE 3000

CMD ["node", "main.js"]