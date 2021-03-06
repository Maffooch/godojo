package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mtesauro/godojo/config"
)

// Commands to bootstrap Ubuntu for the installer
func ubuntuInitOSInst(id string, b *osCmds) {
	switch id {
	case "ubuntu:18.04":
		b.id = "ubuntu:18.04"
		b.cmds = []string{
			fmt.Sprintf("curl -sS %s | apt-key add -", YarnGPG),
			fmt.Sprintf("echo -n %s > /etc/apt/sources.list.d/yarn.list", YarnRepo),
			"DEBIAN_FRONTEND=noninteractive apt-get update",
			"DEBIAN_FRONTEND=noninteractive apt-get install -y apt-transport-https libjpeg-dev gcc libssl-dev python3-dev python3-pip python3-virtualenv nodejs yarn build-essential expect",
		}
		b.errmsg = []string{
			"Unable to obtain the gpg key for Yarn",
			"Unable to add yard repo as an apt source",
			"Unable to update apt database",
			"Installing OS packages with apt failed",
		}
		b.hard = []bool{
			true,
			true,
			true,
			true,
		}
		// Currently, only Ubuntu 18.04 is supported
	}
	return
}

// Commands to install SQLite on Ubuntu
func ubuntuInstSQLite(id string, b *osCmds) {
	switch id {
	case "ubuntu:18.04":
		b.id = id
		b.cmds = []string{
			"DEBIAN_FRONTEND=noninteractive apt-get install -y sqlite3",
		}
		b.errmsg = []string{
			"Unable to install SQLite",
		}
		b.hard = []bool{
			true,
		}
	}
	return
}

// Commands to install MariaDB on Ubuntu
func ubuntuInstMariaDB(id string, b *osCmds) {
	switch id {
	case "ubuntu:18.04":
		b.id = id
		b.cmds = []string{
			"DEBIAN_FRONTEND=noninteractive apt-get install -y mariadb-server libmariadbclient-dev",
		}
		b.errmsg = []string{
			"Unable to install MariaDB",
		}
		b.hard = []bool{
			true,
		}
	}
	return
}

// Commands to install MySQL on Ubuntu
func ubuntuInstMySQL(id string, b *osCmds) {
	switch id {
	case "ubuntu:18.04":
		b.id = id
		b.cmds = []string{
			"DEBIAN_FRONTEND=noninteractive apt-get install -y mysql-server libmysqlclient-dev",
		}
		b.errmsg = []string{
			"Unable to install MySQL",
		}
		b.hard = []bool{
			true,
		}
	}
	return
}

// Commands to install PostgreSQL on Ubuntu
func ubuntuInstPostgreSQL(id string, b *osCmds) {
	switch id {
	case "ubuntu:18.04":
		b.id = id
		b.cmds = []string{
			"DEBIAN_FRONTEND=noninteractive apt-get install -y libpq-dev postgresql postgresql-contrib",
		}
		b.errmsg = []string{
			"Unable to install PostgreSQL",
		}
		b.hard = []bool{
			true,
		}
	}
	return
}

// Determine the default creds for a database freshly installed in Ubuntu
func ubuntuDefaultDBCreds(db string, creds map[string]string) {
	// Installer currently assumes the default DB passwrod handling won't change by release
	// Switch on the DB type
	switch db {
	case "MySQL":
		ubuntuDefaultMySQL(creds)
	}

	return
}

func ubuntuDefaultMySQL(c map[string]string) {
	// Sent some intial values that ensure the connection will fail if the file read fails
	c["user"] = "debian-sys-maint"
	c["pass"] = "FAIL"

	// Pull the debian-sys-maint creds from /etc/mysql/debian.cnf
	f, err := os.Open("/etc/mysql/debian.cnf")
	if err != nil {
		// Exit with error code if we can't read the default creds file
		errorMsg("Unable to read file with defautl credentials, cannot continue")
		os.Exit(1)
	}

	// Create a new buffered reader
	fr := bufio.NewReader(f)

	// Create a scanner to run through the lines of the file
	scanner := bufio.NewScanner(fr)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "password") {
			l := strings.Split(line, "=")
			// Make sure there was a "=" in l
			if len(l) > 1 {
				c["pass"] = strings.Trim(l[1], " ")
				break
			}
		}
	}
	if err = scanner.Err(); err != nil {
		// Exit with error code if we can't scan the default creds file
		errorMsg("Unable to scan file with defautl credentials, cannot continue")
		os.Exit(1)
	}

}

func ubuntuOSPrep(id string, inst *config.InstallConfig, b *osCmds) {
	// Setup virutalenv, setup OS User, and chown DefectDojo app root to the dojo user
	switch id {
	case "ubuntu:18.04":
		b.id = id
		b.cmds = []string{
			"python3 -m virtualenv --python=/usr/bin/python3 " + inst.Root,
			inst.Root + "/bin/pip3 install -r " + inst.Root + "/django-DefectDojo/requirements.txt",
			"mkdir " + inst.Root + "/logs",
			"groupadd " + inst.OS.Group,
			"useradd -s /bin/bash -m -g " + inst.OS.Group + " " + inst.OS.User,
			"chown -R " + inst.OS.User + "." + inst.OS.Group + " " + inst.Root,
		}
		b.errmsg = []string{
			"Unable to setup virtualenv for DefectDojo",
			"Unable to install Python3 modules for DefectDojo",
			"Unable to create a directory for logs",
			"Unable to create a group for DefectDojo OS user",
			"Unable to create an OS user for DefectDojo",
			"Unable to change ownership of the DefectDojo app root directory",
		}
		b.hard = []bool{
			true,
			true,
			true,
			true,
			true,
			true,
		}
	}

	return
}

func ubuntuSetupDDjango(id string, inst *config.InstallConfig, b *osCmds) {
	// Django installs - migrations, create Django superuser
	switch id {
	case "ubuntu:18.04":
		b.id = id
		b.cmds = []string{
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py makemigrations --merge --noinput",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py makemigrations dojo",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py migrate",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py createsuperuser --noinput --username=\"" +
				inst.Admin.User + "\" --email=\"" + inst.Admin.Email + "\"",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && " +
				inst.Root + "/django-DefectDojo/setup/scripts/common/setup-superuser.expect " + inst.Admin.User + " " + inst.Admin.Pass,
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata product_type",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata test_type",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata development_environment",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata system_settings",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata benchmark_type",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata benchmark_category",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata benchmark_requirement",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata language_type",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata objects_review",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata regulation",
			//"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py import_surveys",
			//"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py loaddata initial_surveys",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py buildwatson",
			"cd " + inst.Root + "/django-DefectDojo && source ../bin/activate && python3 manage.py installwatson",
			"cd " + inst.Root + "/django-DefectDojo/components && yarn",
			"cd " + inst.Root + "/django-DefectDojo/ && source ../bin/activate && python3 manage.py collectstatic --noinput",
			"chown -R " + inst.OS.User + "." + inst.OS.Group + " " + inst.Root,
		}
		b.errmsg = []string{
			"Initial makemgrations failed",
			"Failed during makemgration dojo",
			"Failed during database migrate",
			"Failed while creating DefectDojo superuser",
			"Failed while setting the password for the DefectDojo superuser",
			"Failed while the loading data for product_type",
			"Failed while the loading data for test_type",
			"Failed while the loading data for development_environment",
			"Failed while the loading data for system_settings",
			"Failed while the loading data for benchmark_type",
			"Failed while the loading data for benchmark_category",
			"Failed while the loading data for benchmark_requirement",
			"Failed while the loading data for language_type",
			"Failed while the loading data for objects_review",
			"Failed while the loading data for regulation",
			//"Failed while the running import_surveys",
			//"Failed while the loading data for initial_surveys",
			"Failed while the running buildwatson",
			"Failed while the running installwatson",
			"Failed while the running yarn",
			"Failed while the running collectstatic",
			"Unable to change ownership of the DefectDojo directory",
		}
		b.hard = []bool{
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			//true,
			//true,
			true,
			true,
			true,
			true,
			true,
		}
	}

	return
}
