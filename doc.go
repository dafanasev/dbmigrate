package migrate

/*
	Features:
	* up and down migrations
	* different environments (test, dev, prod)
	* specific dialect migrations
	* policies when new migration arrived that has earlier Timestamp when existing - rollback to this Migration or not
	* store migrations in files or embed into binary
	* use as library or CLI tool
	* configuring using flags, yml file or env variables
*/