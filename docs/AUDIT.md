# Gogal Activity Timeline

Gogal records document lifecycle events in `tab_audit_logs` and renders them in the Desk activity panel with comments and assignment placeholders.

## Current Events

- `create`
- `update`
- `delete`
- Desk comments
- Desk assignment placeholders
- Field diff previews for tracked updates

## Actor

The MVP actor is read from request headers:

- `X-Gogal-User`
- `X-User`
- `X-Actor`

If no actor is provided, Gogal stores `system`.

## Timeline

Desk timeline events are served from:

```text
GET /desk/resource/:doctype/:id/timeline
```

Desk activity writes are served from:

```text
POST /desk/resource/:doctype/:id/comment
POST /desk/resource/:doctype/:id/comment/:comment_id/edit
POST /desk/resource/:doctype/:id/comment/:comment_id/delete
POST /desk/resource/:doctype/:id/assign
POST /desk/resource/:doctype/:id/assignment/:assignment_id/open
POST /desk/resource/:doctype/:id/assignment/:assignment_id/closed
```

## Tables

- `tab_audit_logs`: immutable lifecycle events.
- `tab_comments`: simple document comments.
- `tab_assignments`: open assignment placeholders for future users, permissions, workflow, and notifications.

## Track Changes

DocTypes control lifecycle audit events with `track_changes`.

When `track_changes` is false, create/update/delete mutations still work, but Gogal does not write lifecycle audit rows for that DocType.

Comments and assignments remain explicit Desk activity, so they continue to appear in the timeline.

## Diff Preview

Tracked updates store a small JSON diff preview in audit metadata:

```json
{
  "diffs": [
    {"field": "customer_name", "before": "Old", "after": "New"}
  ]
}
```

The timeline renders this as a compact field-level preview. It is not yet a full document versioning system.

The timeline is intentionally small and non-blocking. Future workflow, permissions, notifications, and full version history can extend this foundation.
