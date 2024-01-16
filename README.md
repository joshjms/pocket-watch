# Pocket Watch

REST API for running unreliable code in a sandboxed environment.

> The service is still under development. 

## Usage

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

## Contributing
TBD - For more information, contact me in Discord `joshjms`. 

## People
* [Josh](https://github.com/joshjms)
* [Clay](https://github.com/sanstzu)
