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
  -f string
        format (default "text")
  -m string
        name of mode (default "cli")
  -s    enable streaming
```

### Available drivers
- GitHub Events
- GitHub Issues
- Gmail
- Tumblr
- Twitter
- Reddit

## Example
### Stream posts of Twitter and Reddit
```
smoothie -s twitter reddit
```