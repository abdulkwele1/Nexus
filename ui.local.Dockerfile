FROM node:lts-alpine

WORKDIR /app

EXPOSE 5173
CMD npm install && npm run dev
