<!--
Templates:

## Unreleased

#### Features

#### Improvements

#### Bug Fixes

#### Breaking changes
-->

## Unreleased

#### Features

#### Improvements
- (deps) [#41](https://github.com/bcdevtools/consvp/pull/41) Update deps cvp-streaming-core v1.1.0
- (codec) [#42](https://github.com/bcdevtools/consvp/pull/42) Add new flag `--codec` to specify codec for streaming
- (errors) [#43](https://github.com/bcdevtools/consvp/pull/43) Queue message and print after T-UI closed to prevent message lost & improve error message content
- (errors) [#44](https://github.com/bcdevtools/consvp/pull/44) Minimize error message content
- (broadcast) [#45](https://github.com/bcdevtools/consvp/pull/45) Stop app when broadcast failed with response http status 404

#### Bug Fixes

#### Breaking changes

## Release v1.0.3

#### Improvements
- (broadcast) [#38](https://github.com/bcdevtools/consvp/pull/38) Improve message content when server returns 404 as result of broadcast
- (exit) [#39](https://github.com/bcdevtools/consvp/pull/39) Execute cleanup methods gracefully when app exit

## Release v1.0.2

#### Improvements
- (docs) [#35](https://github.com/bcdevtools/consvp/pull/35) Add CHANGELOG.md
- (make) [#36](https://github.com/bcdevtools/consvp/pull/36) Use git tag to mark binary version during `make install` or `make build`

#### Bug Fixes
- (broadcast) [#34](https://github.com/bcdevtools/consvp/pull/34) Reflect fetch issue in broadcast paragraph T-UI

## Release v1.0.1

#### Bug Fixes
- (update) [#31](https://github.com/bcdevtools/consvp/pull/31) Remove command flag `--update` and update docs

## Release v1.0.0

#### Features
- Full rework of [pvtop](https://github.com/blockpane/pvtop)
- Live-streaming mode ([view sample](https://cvp.bcdev.tools/pvtop/sample-chain-1_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA))
- Display Block Hash fingerprint which the validator voted on
- Allow scrolling on terminal UI