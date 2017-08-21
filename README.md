# hana
Data point connecter to prometheus for any data source

### build

	make
	make linux

Above commands will create a single binary in `build` folder.

	make test

Above command will run unit tests

### usage

	./build/hana -d conf/asaka.conf
