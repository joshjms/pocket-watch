# Pocket Watch

REST + gRPC API for running unreliable code in a sandboxed environment. Everyone can make their very own Leetcode without compromising server security and health.

> The service is still under development. Please ⭐ this project to motivate the coding monkeys ❤️!

## Usage

### Initialization

#### Docker

```bash
docker run -p 8080:8080 \
--privileged --cap-add SYS_ADMIN --security-opt seccomp=unconfined --cgroup-parent=docker.slice --cgroupns private --tmpfs /tmp --tmpfs /run --tmpfs /run/lock \
pocketwatch:latest
```

### Running Source Code (C++ Example)

> Currently, only C/C++ source codes are supported. Add support for other languages in the future.

POST `/run`

#### Request 
```go
{
	Code     string   `json:"code"`
	Language string   `json:"language"` // just use "cpp"; other languages not yet supported
	Stdin    []string `json:"stdin"`
}
```

#### Response
```go
{
	Verdict []string `json:"verdict"` // OK, RE, TLE, CE
	Stdout  []string `json:"stdout"`
	Stderr  []string `json:"stderr"`
	Time    []string `json:"time"`
	Memory  []string `json:"memory"`
}

```

## Contributing
TBD - For more information, contact me in Discord `joshjms`.

## References
- [pocket-api](https://github.com/sanstzu/pocket-api)

## People
* [Josh](https://github.com/joshjms)
* [Clay](https://github.com/sanstzu)

