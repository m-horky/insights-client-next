# Working on Insights Client

## Project structure

The repository aims to be both CLI application and a Go library usable by other programs.

Every module MUST return `internal.IError`-compatible errors to allow for easy diagnosis and presentation of the application errors.

- `api/`

  Module containing methods and objects required to talk to Insights APIs.

- `api/ingress/`, `api/inventory/`, ...

  Modules for specific Insights services Insights Client talks to.
  Its scope is very limited, only API endpoints required by the application are implemented.

  Since the host configuration is out of scope for this part of the library, each service has to be configured from external code:

  ```go
  package main
  import (
      "github.com/.../insights-client/api"
      "github.com/.../insights-client/api/inventory"
  )
  func init() {
      url := api.NewServiceURL("https", "cert.console.redhat.com", "443")
      inventory.Init(api.NewServiceWithAuthentication(url, "/path/to/cert.pem", "/path/to/key.pem"))
  }
  ```
  
  Each API defines its specific errors; they are always prefixed with `Err` (e.g. `ErrNoHost` returned by Inventory, or `ErrArchive` returned by Ingress-related methods).

- `cmd/`

  Source code for the CLI itself.

- `modules/`

  Module managing communication with data collectors.

- `internal/`

  Sources for the behavior of CLI.
