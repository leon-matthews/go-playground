## Development Database

Database for development only:

## Backup

```bash
sudo mariadb-dump --single-transaction --routines --events --databases snippetbox > database.sql
```

## Restore

```bash
sudo mariadb snippetbox < backup.sql
```
