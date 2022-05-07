# Schedule Parser

Schedule Parser is a library for parsing pdf schedules with specific layout.

## Installation

```
go get -u github.com/qsoulior/scheduleparser
```

## Features

### Parse file

```go

initialDate := time.Now()

err := scheduleparser.ParseFile("input.pdf", "output.json", initialDate)
if err != nil {
  log.Fatal(err)
}
```

### Parse bytes

```go
initialDate := time.Now()

// os.ReadFile (as of Go 1.16) reads file and returns bytes
contents, err := os.ReadFile("input.pdf")
if err != nil {
  log.Fatal(err)
}
// or get bytes in another way 

result, err := scheduleparser.ParseBytes(contents, initialDate)
if err != nil {
  log.Fatal(err)
}
```
