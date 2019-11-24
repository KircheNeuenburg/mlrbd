package main

import (
	"database/sql"
	"fmt"
	"log"
	//	"maunium.net/go/mautrix"
	"mlrbd/config"
	"mlrbd/db"
	"mlrbd/ldap"
	"mlrbd/matrix"
)

var (
	dbc  *sql.DB
	conf *config.Config
)

func main() {
	var err error
	conf, err = config.LoadConfigFromFile(nil, "")
	dbc, err = db.NewConnectionPool(
		conf.Db.Connection,
		1,
		5,
	)
	if err != nil {
		log.Fatal("Unable to connect to the database: %v", err)
	}
	defer dbc.Close()

	matrix.StartMatrix(conf)
	ldap.StartLDAP(conf)
	lg_act, err := ldap.GetLDAPGroups()
	if err != nil {
		log.Fatal(err)
	}
	lg_curr, err := db.GetDbGroups(dbc)
	if err != nil {
		log.Fatal(err)
	}

	handleCreatedGroups(Diff(lg_curr, lg_act))
	handleCurrentGroups(Intersect(lg_curr, lg_act))
	handleRemovedGroups(Diff(lg_act, lg_curr))

	mr_act, err := db.GetDbRooms(dbc)
	if err != nil {
		log.Fatal(err)
	}
	mr_curr, _ := matrix.GetMatrixRooms()
	handleCleanedRooms(Diff(mr_act, mr_curr))
	ldap.StopLDAP()
}

func Intersect(a []string, b []string) (set []string) {
	hash := make(map[string]bool)

	for _, s := range a {
		hash[s] = true
	}

	for _, s := range b {
		if _, found := hash[s]; found {
			set = append(set, s)
		}
	}

	return
}

func Diff(a []string, b []string) (diff []string) {
	hash := make(map[string]bool)

	for _, s := range a {
		hash[s] = true
	}

	for _, s := range b {
		if _, found := hash[s]; !found {
			diff = append(diff, s)
		}
	}

	return
}

func handleRemovedGroups(lg []string) {
	stmt_sel, err := dbc.Prepare("SELECT matrix_room FROM group_mapping WHERE ldap_group = $1 LIMIT 1")
	defer stmt_sel.Close()
	if err != nil {
		log.Fatal(err)
	}
	stmt_del, err := dbc.Prepare("DELETE FROM group_mapping WHERE ldap_group = $1")
	defer stmt_del.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range lg {
		fmt.Println("Remove Group ", s)
		row := stmt_sel.QueryRow(s)
		var rid string
		if err := row.Scan(&rid); err != nil {
			log.Fatal(err)
		}
		matrix.DeleteMatrixRoom(rid)
		_, err = stmt_del.Exec(s)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handleCleanedRooms(mr []string) {
	for _, rid := range mr {
		fmt.Println("Cleanup Room ", rid)
		matrix.DeleteMatrixRoom(rid)
	}
}

func handleCreatedGroups(lg []string) {
	stmt, err := dbc.Prepare("INSERT INTO group_mapping(ldap_group,matrix_room) VALUES($1, $2)")
	defer stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range lg {
		n := ldap.GetLdapGroupName(s)
		rid := matrix.CreateMatrixRoom(n)
		fmt.Println("Create Group ", n)
		if _, err := stmt.Exec(s, rid); err != nil {
			log.Fatal(err)
		}
		lu, err := convertToMxid(ldap.GetLdapUsers(s))
		if err != nil {
			log.Fatal(err)
		}
		matrix.HandleCreatedUsers(rid, lu)
	}
}

func handleCurrentGroups(lg []string) {
	stmt, err := dbc.Prepare("SELECT matrix_room FROM group_mapping WHERE ldap_group = $1 LIMIT 1")
	defer stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Syncing Groups")
	for _, s := range lg {
		row := stmt.QueryRow(s)
		var rid string
		if err := row.Scan(&rid); err != nil {
			log.Fatal(err)
		}
		lu, err := convertToMxid(ldap.GetLdapUsers(s))
		if err != nil {
			log.Fatal(err)
		}
		mu := matrix.GetMatrixUsers(rid)

		matrix.HandleCreatedUsers(rid, Diff(mu, lu))
		matrix.HandleRemovedUsers(rid, Diff(lu, mu))
	}
}

func convertToMxid(lu []string, err_in error) (mu []string, err error) {
	err = err_in
	for _, e := range lu {
		mu = append(mu, "@"+e+":"+conf.Matrix.Homeserver)
	}
	return
}
