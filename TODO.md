# Things to do

## Source code maintaining

- [ ] deal with TODO's in every source file
- [ ] add an order of initialization resolution

## Documenting

- [ ] write a more precise documentation to every type, function, method, and
      file (it should be `godoc`able)
- [ ] write a *Symbols Extractor Design & Hacking Guide*
- [ ] give a proposal of doc comment (`godoc` friendly, of course)
- [ ] target on interactive generated documentation
- [ ] in `docs/hacking.md`, write more detailed info about each type of
      generator

## Testing

- [ ] write tests that covers all *Symbol Extractor* requirements
- [ ] write tests for data types propagation
- [ ] write a test generator in `golang & YAML` that generates complete test
      suite and its documentation
      * `strict` mode enabled by `--strict` or `--mode=strict` flags checks
        for reduntand items in YAML files
- [ ] test generator should be the part of `update` commands
