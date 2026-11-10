# Omisha OS 5.2 — PostgreSQL Backup and Restore Smoke Test

These scripts operate only on the `scale/` PostgreSQL database. They do **not** back up customer Docker volumes, object storage, the Lite Core JSON registry, or licensing signing keys.

## Backup
```bash
bash scale/backup.sh /secure/omisha-backups
```
Backups are compressed, created with restricted permissions, and accompanied by SHA-256 checksums. Store copies off-host in encrypted storage with retention rules.

## Restore smoke test
```bash
bash scale/restore-test.sh /secure/omisha-backups/omisha-TIMESTAMP.sql.gz
```
This creates a *new* temporary database and imports the dump; it does not overwrite the live `omisha` database. It does not automatically delete the temporary database.

**Not production-ready:** No automated schedule, encryption-at-rest management, offsite replication, PITR, backup verification alerts, or recovery time objective measurements. Always test backups and restores before relying on them.
