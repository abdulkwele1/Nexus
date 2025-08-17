# Developer Operations

## Goals

1. Hosting - When user's go to https://nexus.eternalrelayrace.com, they are able to access the latest version of the app
2. Deployment - Developers should be able to deploy new versions of the app after code is reviewed and merged into the main branch

### Hosting Steps

1.) Acquire domain name (e.g. from https://www.hover.com), this is the "address of your business" - e.g. the Nexus website, on "the internet". When people want to "see you", they go to your address - https://nexus.eternalrelayrace.com
2.) Acquire a public IP (Internet Protocol) address - this is the address of the computer (host) that your website runs on. Example IPv4 address - 35.94.111.25
3.) Add DNS record to your domain name to map it to the ip address of the host your website and api runs on
4.) Turn on your computer / launch a cloud instance - for example AWS EC2
5.) Associate your IP address with your computer

Now whenever a user goes to https://nexus.eternalrelayrace.com, their browser will look up the DNS entry for that website, receive an answer of 35.94.111.25, and send the users traffic to the computer with that address on the internet.

### Deployment Steps

1.) Download the private ssh key to the server (ask Abdul or Levi)
2.) Set the required security permissons for the key `chmod 400 nexus-demo-server.pem`
3.) SSH (remote login) to the server `ssh -i nexus-demo-server.pem ubuntu@35.94.111.25` (type yes when it asks you if you want to trust this server)
4.) `cd Nexus` to go to the Nexus code repo
5.) Get the latest merged code `git pull -p`
6.) If you need to deploy updates to the api run `sudo make refresh`
7.) If you need to seed the database run `sudo make seed-database`
8.) If you need to deploy updates to the frontend run `sudo docker compose up -d  nexus-ui --build --force-recreate`

##### Operations

If you need to debug and view the logs run `sudo make logs`

If you need to debug the state of the database `sudo make debug-database`
