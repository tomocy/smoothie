# smoothie
client for various social media

## Installtion
```
go install github.com/tomocy/smoothie/cmd/smoothie
```

## Prerequisites
- Register an app in social media you want to use as the drivers and Keep the Client ID and Secert
- Set the IDs and Secrets in your env or in .env file located anywhere (default: current directory) ([.env example](.env.example))

## Usage
```
Usage of smoothie: [optinos] drivers...
  -env string
        the path to .env (default "./.env")
  -f string
        format (default "text")
  -m string
        name of mode (default "cli")
  -v string
        verb (default "fetch")
```

### Available drivers
- github:events
- github:issues
- gmail
- tumblr
- twitter
- reddit

## Example
### Stream posts of Twitter and Reddit
```
smoothie -v stream twitter reddit
```