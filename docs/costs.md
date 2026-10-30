# Costs

Rudolph on Google Cloud bills for the load balancer, Cloud Run, Firestore, and, when event export is on, Pub/Sub, Cloud Storage, and Cloud KMS. There is no per-request API Gateway charge and no DynamoDB read-capacity unit.

What dominates depends on how often Santa syncs:

- **Cloud Run** CPU and memory while a request is being handled, plus always-on cost if `min_instances` is greater than zero. The default is zero instances, so an idle fleet does not keep a container hot. A minimum of one instance removes the cold start on the first sync after idle.
- **Firestore** document reads and writes. Preflight writes sensor data and sync state and reads configuration. Ruledownload reads a page of rules. Postflight updates sync state and deletes stale machine rules. Global rules change only when an operator uses the CLI.
- **External Application Load Balancer** forwarding rules and outbound data. The managed certificate and the trust config are part of Certificate Manager.
- **Pub/Sub and Cloud Storage** only for execution events. The archive bucket deletes objects after `event_retention_days` (default 365). Turning `enable_event_export` off removes the topic, the subscription, and the bucket.
- **Cloud KMS** for the key that encrypts the events bucket. A key with a 90 day rotation is created even when you later disable event export, because the key ring is shared. Delete the stack if you do not want the key.

The old AWS estimate (on the order of hundreds of dollars a month for thousands of clients syncing every minute) does not transfer. Use the Cloud Billing report filtered to this project. Syncing every 10 minutes, which is Santa's usual default, does far less work than syncing every minute.

Rule count affects ruledownload reads during a clean sync, because a clean sync pages through every global rule. Steady-state incremental syncs read the feed since the cursor, which is usually small.
