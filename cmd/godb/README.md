# godb CLI

```sh
go install github.com/fzf-labs/godb/cmd/godb@latest
```

`godb` provides four commands:

```sh
godb ormgen     # Generate GORM model/dao/repo code
godb sqldump    # Export table schema SQL only
godb sqlbackup  # Export one database's schema and data
godb sqltopb    # Generate proto files from table schemas
```

`sqlbackup` creates a PostgreSQL custom archive or a MySQL SQL dump. It is not a restore command; use `pg_restore` or `mysql` to restore the artifact. See the repository [README](../../README.md) for command flags, examples, prerequisites, and restore instructions.
