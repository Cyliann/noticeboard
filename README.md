# Noticeboard
App created as a passion project to learn the inner workings of Go and backend programming.
It's purpose was to serve as a downdetector for my company, but was never used, so it lands in my portfolio.

## Supports:
- viewing announcements as a viewer
- logging in
- adding, removing and editing announcements as an admin
- password change
- full API documentation using [Swagger](https://swagger.io/) 

## Stack:
- SQLite as a DB
- [Gorilla sessions](https://github.com/gorilla/sessions) for cookie management
- [Bootstrap v.5](https://github.com/twbs/bootstrap) for frontend

## How to run:
`go run .` in the project directory or build the project using `go build .` and run the resulting binary. 

By default the app listens on port `8080`.

If there is no `reports.db` SQLite file, app will create it on its own.

The documentation is available on `localhost:8080/docs/`

There is a default user with credentials "admin:changeme" for testing purposes.
