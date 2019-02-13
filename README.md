Browser Forensics
==========

This project exists as a simple multi-platform application to:
  - Read/Write the following through a common interface
    - History
    - Cookies
    - Credentials (limited)
    - Bookmarks
  - Backup data to a secure common interface
  - Restore a set of common interface data into a given browser  

Development
===========

This project is developed using Docker.

### Project Structure
The project is follows the established paradigms in [golang standards project-layout](https://github.com/golang-standards/project-layout) project and is broken down into the following directory structure:

```
├── cmd/
│   ├── main.go
│   ├── ...
├── configs/
│   ├── ...
├── dockerfiles/
│   ├── main.go
│   ├── ...
├── pkg/
│   ├── browsers
│   ├── ...
├── scripts/
│   ├── build.sh
│   ├── ...
...
```


* `cmd/` Main applications for this project (at this time just one, more to come)
* `configs/` Configuration file templates or default configs (hard coded and compiled for the moment)
* `dockerfiles/` Docker files describing build and run containers
* `pkg/` Library code that's ok to use by external applications
* `scripts/` Scripts to perform various build, install, analysis, etc operations

### Idioms
* `go fmt`
* Whole-word variable names
* Don't log, return errors and expect consumer handling

Build
===========
To build binaries for all supported platforms:

```
$ scripts/build.sh
```

How to use
===========

### Dependencies
This project is statically compiled Go / CGo code (except for macOS) and should have no dependencies to run.

### Run
Port the correct binary to your operating environment and run it through the terminal or by double clicking.