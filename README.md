# smoothie
client for various social media

## Installtion
```
go install github.com/tomocy/smoothie/cmd/smoothie
```

## Prerequisites
- Register an app in the development console of the social media you want to use as drivers and keep the Client ID and Secert
- Set the IDs and the Secrets in your env or in .env file located anywhere (default: .env in current directory is used) and name them as `{driver name}_CLIENT_ID` and `{driver name}_CLIENT_SECRET` ([.env example](.env.example))

## Example
- fetch GitHub issues of [golang/go](https://github.com/golang/go)
```
smoothie github:issues:golang/go
```
- stream posts of Twitter and Reddit
```
smoothie -v stream twitter reddit
```

## Usage
```
Usage of smoothie: [optinos] drivers...
  -env string
        the path to .env (default "./.env")
  -f string
        format (default "text")
  -m string
        mode (default "cli")
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

### Avaiable args
- github:events
```
github:events:{username}
```
- github:issues
```
github:issues:{owner/repo}
```
