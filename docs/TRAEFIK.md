# Gogal Traefik Plan

Traefik integration is optional and not required for MVP.

## Future Goal
Site/domain routing patterns such as:
- `site1.localhost`
- `company1.gogal.local`
- `abc123.gogal.cloud`
- custom domains

## Placeholder Packages
- `internal/proxy/traefik.go`
- `internal/proxy/domain.go`

## Future CLI (Planned)
- `gogal proxy status`
- `gogal proxy setup-traefik`
- `gogal proxy add-site --site demo.local --domain demo.localhost`
- `gogal proxy random-domain --site demo.local`

## MVP Behavior
- `gogal start` is not blocked by Traefik availability.
- `gogal doctor` can later show Traefik as optional status.
