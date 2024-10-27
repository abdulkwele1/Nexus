# Nexus

This is the monorepo of front end clients and backend services for collecting, processing, viewing and analyzing energy data related to [Black Futures Farm](https://blackfutures.farm).

## Using Nexus

### Create Accounts

In the future we plan to add the ability for admin users to create new user accounts, for now this is a manual process that involves

1.) SSH'ing to the Nexus server

```bash
ssh -i nexus-demo-server.pem ubuntu@35.94.111.25
```

2.) Changing to the sudo user

```bash
ubuntu@ip-172-31-29-168:~$ sudo -s
root@ip-172-31-29-168:/home/ubuntu#
```

3.) Changing directory to the Nexus code directory

```bash
root@ip-172-31-29-168:/home/ubuntu# cd Nexus/
root@ip-172-31-29-168:/home/ubuntu/Nexus#
```

4.) Connecting to the database

```bash
root@ip-172-31-29-168:/home/ubuntu/Nexus# make debug-database
docker compose exec nexus-db psql -U postgres -d postgres
psql (16.4 (Debian 16.4-1.pgdg120+2))
Type "help" for help.

postgres=#
```

5.) Inserting a new row into the login_authentications table with the name of the user and hash of default password `password123`

```bash
postgres=# INSERT INTO "login_authentications"
 ("id", "user_name", "password_hash")
 VALUES
 (DEFAULT, 'levi', '$2a$10$HqQx4jxUzfQm1fZYUZRLbOBaMNWHmhSmweH03rl0EykgE4BNfDciO')
  ON CONFLICT ("user_name") DO NOTHING;
INSERT 1 0
postgres=# quit
root@ip-172-31-29-168:/home/ubuntu/Nexus#
```

Lastly share the user_name and default password `password123` with the new user.

### Login to Your Account

1.) Navigate to <https://nexus.eternalrelayrace.com>
2.) Enter your username and password
3.) Click login

### Viewing Solar Data

- add details about how to navigate to solar data view
- add details that explain what the metrics of solar yield and solar consumption are, e.g. what do they measure
- add details about how to select time range
- add details about how to export data

## Developing Nexus

[DEVELOPMENT.md](./DEVELOPMENT.md)
