package migrate

/*
	Features:
	* up and down migrations +
	* isSpecific dialect migrations +
	* store migrations in files +
	* or embed into binary
	* use as library +
	* or CLI tool
	* migrations generator +
	* printing progress (using migration and error channels) +
	* option to return error or skip step on rollback and down migration is not exist or empty
	* option to return error on subfolder or wrong file name in the migrations folder
	* policies when new migration arrived that has earlier Timestamp when existing - rollback to this migration or not
	* configuring using flags, yml file or env variables
	* different environments (e.g. test, dev, prod)
	* move insert/delete versions logic into transaction

	TODO:
	* Dockerfile
	* comments
	* Readme
*/
