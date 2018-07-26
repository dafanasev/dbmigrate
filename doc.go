package migrate

/*
	Features:
	* up and down migrations +
	* specific dialect migrations +
	* store migrations in files +
	* or embed into binary
	* use as library +
	* or CLI tool
	* printing progress (using migration and error channels)
	* option to return error or skip step when rollback and down migration is not exist
	* graceful shutdown
	* configuring using flags, yml file or env variables
	* different environments (e.g. test, dev, prod)
	* policies when new migration arrived that has earlier Timestamp when existing - rollback to this Migration or not

	TODO:
	* Dockerfile
	* Vendor libs
*/