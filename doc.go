package migrate

/*
	Features:
	* up and down migrations
	* different environments (test, dev, prod)
	* specific dialect migrations
	* policies when new migration arrived that has earlier timestamp when existing - rollback to this migration or not
	* store migrations in files or embed into binary
	* use as library or CLI tool
	* configuring using flags, yml file or env variables
*/