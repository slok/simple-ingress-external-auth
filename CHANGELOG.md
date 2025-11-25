# Changelog

## [Unreleased]

## [v0.6.0] - 2025-11-25

### Changed

- Dependency upgrades

## [v0.5.0] - 2025-04-30

### Changed

- Disabled tokens now will act as missing/unknown tokens, they will not return "disabledToken" error nor metric.
- Dependency upgrades

## [v0.4.0] - 2022-12-10

### Added

- `client_id` now will be returned as a header, by default in `X-Ext-Auth-Client-Id` header.
- Add `--client-id-header` CMD flag to customize the header that will be returned with the authenticated client.
- Add `--request-method-header` cmd flag to customize the header that will be checked to get the original request method, by default the one that Nginx uses: `X-Original-Method`.
- Add `--request-url-header` cmd flag to customize the header that will be checked to get the original request method, by default the one that Nginx uses: `X-Original-URL`.

## [v0.3.0] - 2022-08-01

### Added

- `client_id` option on tokens to identify the tokens.
- `client_id` will be shown on token review metrics.

## [v0.2.0] - 2022-06-26

### Added

- HTTP metrics.

## [v0.1.0] - 2022-06-21

### Added

- Token configuration.
- JSON configuration support.
- YAML configuration support.
- HTTP handler for authentication.
- Add token review Prometheus metrics.
- Make configuration API public as a go library.

[unreleased]: https://github.com/slok/simple-ingress-external-auth/compare/v0.6.0...HEAD
[v0.6.0]: https://github.com/slok/simple-ingress-external-auth/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/slok/simple-ingress-external-auth/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/slok/simple-ingress-external-auth/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/slok/simple-ingress-external-auth/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/slok/simple-ingress-external-auth/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/slok/simple-ingress-external-auth/releases/tag/v0.1.0
