# Handlers

HTTP handlers for the Santa sync API. Each handler decides whether it owns the route template, boots its Firestore client once, and returns a JSON response. `cmd/api` serves them on `$PORT`.
