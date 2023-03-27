# Contribution guideline

## Issue reporting

If you encounter any problem while working with `scotty` feel free to open a new issue. Ensure that the issue includes

- ..the effected version (`scotty version`)
- ..a step by step guide of to reproduce the issue alongside context about the kind of logs streamed

## Feature request

Something is missing for in scotty? Make a proposal of your idea by creating an [issue](https://github.com/KonstantinGasser/scotty/issues/new?assignees=&labels=&template=feature_request.md&title=).

Try to explain why you think this feature could be useful also to others by solving which problem?

## Code contributions

**note: the code contribution guidelines also apply the [beam](https://github.com/KonstantinGasser/beam) repository.
however please refer to issues and contributions within its repository**

### Local environment

Developing with go-1.18 for potential usage of generics.

Within the scotty repository there is a `/test` directory with a dummy application producting structured logs. This can be used for testing. With the flag `-delay` one can increase the random time-out between each logs which can be useful to either stress test a feature or slow down for debugging.

### How to contribute

1. Start a dialog on a ticket you want to solve, and discuss any open questions closing with a statement to work on the issue
2. Fork the repository & clone the repository
3. verify all works with `go run -race main.go version`

### Conventions

- Committed code should follow the [Effective Go](https://go.dev/doc/effective_go) guidelines.
- Code should be formatted with `gofmt -s` and checked with `go vet`.
- Types and functions exported by a package should have a comment stating what they are doing.

### Contributing changes

1. implement all required changes to solve the ticket and write tests if applicable and useful.
2. Test your changes locally and ensure all other tests continue to pass using `go test -v ./...`.
3. Commit your changes to your forked repository and open a PR with a reference to the issue/ticket.
