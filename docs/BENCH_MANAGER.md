# Gogal Bench Manager

Route: `GET /bench`

## Purpose
Bench-level UI control panel for Gogal operations.

## MVP Skeleton (Current)
- Bench status summary
- Detected sites
- Detected installed apps
- Migration/PostgreSQL/port placeholders
- Disabled action buttons:
  - New Site
  - Install App
  - Run Migration Plan
  - Backup
  - Restore
  - Add Domain
  - Generate Random Subdomain

## Security Note
Bench Manager must not execute dangerous shell commands directly from unauthenticated UI. Access should be restricted to System Manager/Admin in future auth layer.
