package migrate

/*
	Features:
	* up and down migrations +
	* isSpecific dialect migrations +
	* store migrations in files +
	* use as library +
	* migrations generator +
	* printing progress (using migration and error channels) +
	* option to return error or skip step on rollback and down migration is not exist or empty +
	* move insert/delete versions logic into transaction
	* policies when new migration arrived that has earlier Timestamp when existing one - rollback to this migration before migrate new ones or not
	* CLI tool
	* configuring using flags, yml file or env variables
	* different environments (e.g. test, dev, prod)
	* embed migrations into binary


	TODO:
	* Dockerfile
	* comments
	* Readme
*/
