dev_db = tasks_dev
dev_db_user = tasks_dev

.PHONY: psql
psql:
	psql -U $(dev_db_user) $(dev_db)

.PHONY: createdbuser
createdbuser:
	sudo -u postgres createuser -U postgres $(dev_db_user) --createdb --no-password

.PHONY: createdb
createdb:
	createdb -U $(dev_db_user) $(dev_db)
	psql -U $(dev_db_user) $(dev_db) -f tasks.sql

.PHONY: dropdb
dropdb:
	dropdb -U $(dev_db_user) $(dev_db)
