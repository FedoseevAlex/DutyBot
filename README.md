# dutybot
Telegram bot assign a duty + reminder.

# Database setup
Launch postgres and do the following:
- Create database
- Create RW user for application and grant all permissions for the database from previous step
- Enable uuid-ossp extention to postgres

For examples see `internal/migrations/init.sql`.

# Launch app
There is two launch options: direct and run in container
## Run directly
Populate .env file with configuration. 
Current configuration variables is in `internal/config/config.go`.
Example .env file file for fish shell:
```shell
# Database connect string format:
# postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]
set -Ux DB_CONNECT_STRING 'postgresql://postgres:mysecretpassword@127.0.0.1:5432/<your db name>?sslmode=disable'
# To create a bot one can use botfather: https://telegram.me/BotFather
set -Ux BOT_TOKEN '<your telegram bot token>'
# Never assign listen address to 0.0.0.0 in production environment. This is just an example.
set -Ux LISTEN_ADDRESS 0.0.0.0:8080
```
To enable configuration you just need to source it:
```shell
source .env
```

Finally let's lauch app:
```shell
go run cmd/dutybot/main.go
```

## Run in container
### Build container 
That's pretty simple.
```shell
docker build -t dutybot:latest .
```
or
```shell
docker buildx build --platform linux/amd64 -t fedoseevalex/dutybot:latest .
```
and finally (if you have permission to push)
```shell
docker push fedoseevalex/dutybot:latest
```

### Pull from docker hub and run
! Not tested these commands myself, so please adjust !
```shell
docker pull fedoseevalex/dutybot:latest
docker run --env-file .env --expose 8443:8433 --name dutybot --detach fedoseevalex/dutybot:latest 
```

# How to make self signed certificate for bot
Original instruction: https://core.telegram.org/bots/self-signed
Create keys first
```shell
openssl req -newkey rsa:2048 -sha256 -nodes -keyout <private key name>.key -x509 -days 365 -out <public key name>.pem -subj "/C=<Country code>/ST=<city>/L=<location>/O=<organization>/CN=<server ip>"
```

And then inform telegram that you will use self signed cert
```shell
curl -F "url=https://<server ip>:<hook port>/<bot token>" -F "certificate=@<public key name>.pem" <https://api.telegram.org/bot><bot token>/setWebhook
```

