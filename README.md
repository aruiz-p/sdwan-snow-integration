# sdwan-snow-integration

This repo is meant to give guidance while integrating Cisco SD-WAN Manager with Service Now to open and close incidents automatically using wehbooks. You can find the post I created about it [**here**](https://netwithalex.blog/tracking-sd-wan-incidents-with-service-now/) to get a full explanation not only of the code, but the SD-WAN configuration and a little demo. 

# Components

1. SD-WAN enviornment with wan edges deployed. SD-WAN Manager 20.12, lower versions can be used as well.
2. Service Now instance, you can request a developer instance for free.
3. Webhook app built on Golang

# Running the Webhook Server 

1. Replace the envarionment variables with your Service Now instance, user and password.

```bash
SNOW_USER="<YOUR_USER>"
SNOW_PASS='<YOUR_PASSWORD>'
SNOW_INSTANCE="<YOUR_INSTANCE>"
```
2. Run the app server from the root of the repository 

```bash
go mod download
go run ./*.go
```
3. Trigger a webhook event from vManage. See images [**here**](https://netwithalex.blog/tracking-sd-wan-incidents-with-service-now/)


