# smoothie
client for various social media

## Installtion
```
go install github.com/tomocy/smoothie/cmd/smoothie
```

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
- Tumblr
- Twitter
- Reddit

## Example
### Stream posts of Twitter and Reddit
```
smoothie -s twitter reddit
```