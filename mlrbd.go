package main

import (
	"database/sql"
	"log"
	"mlrbd/config"
	"mlrbd/database"
	"mlrbd/ldap"
	"mlrbd/matrix"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	db   *sql.DB
	conf *config.Config
)

func main() {
	var err error
	conf, err = config.LoadConfigFromFile(nil, "")
	db, err = database.NewConnectionPool(
		conf.Db.Connection,
		1,
		5,
	)
	if err != nil {
		log.Fatal("Unable to connect to the database: %v", err)
	}
	defer db.Close()

	database.Migrate(db)

	matrix.StartMatrix(conf)
	ldap.StartLDAP(conf)

	sync()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	ticker := time.NewTicker(time.Duration(conf.General.SyncInterval) * time.Minute)
	done := make(chan bool)
	endsync := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				endsync <- true
				return
			case <-ticker.C:
				sync()
			}
		}
	}()

	<-stop
	ticker.Stop()
	done <- true
	<-endsync
	log.Println("Ticker stopped")
	ldap.StopLDAP()
}

func sync() {
	lg_act, err := ldap.LDAPGroups()
	if err != nil {
		log.Fatal(err)
	}
	lg_curr, err := database.DbGroups(db)
	if err != nil {
		log.Fatal(err)
	}

	handleCreatedGroups(Diff(lg_curr, lg_act))
	handleCurrentGroups(Intersect(lg_curr, lg_act))
	handleRemovedGroups(Diff(lg_act, lg_curr))

	if conf.General.KeepRooms == false {
		mr_act, err := database.DbRooms(db)
		if err != nil {
			log.Fatal(err)
		}
		mr_curr, _ := matrix.MatrixRooms()
		handleCleanedRooms(Diff(mr_act, mr_curr))
	}
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
	stmt_sel, err := db.Prepare("SELECT matrix_room FROM room_group_map WHERE ldap_group = $1 LIMIT 1")
	defer stmt_sel.Close()
	if err != nil {
		log.Fatal(err)
	}
	stmt_del, err := db.Prepare("DELETE FROM room_group_map WHERE ldap_group = $1")
	defer stmt_del.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range lg {
		log.Println("Remove Group ", s)
		row := stmt_sel.QueryRow(s)
		var rid string
		if err := row.Scan(&rid); err != nil {
			log.Fatal(err)
		}
		if conf.General.KeepRooms == false {
			log.Println("Do not keep room ", rid)
			matrix.DeleteMatrixRoom(rid)
		}
		_, err = stmt_del.Exec(s)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handleCleanedRooms(mr []string) {
	for _, rid := range mr {
		log.Println("Cleanup Room ", rid)
		matrix.DeleteMatrixRoom(rid)
	}
}

func handleCreatedGroups(lg []string) {
	stmt, err := db.Prepare("INSERT INTO room_group_map(ldap_group,matrix_room) VALUES($1, $2)")
	defer stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range lg {
		n := ldap.LdapGroupName(s)
		rid := matrix.CreateMatrixRoom(n)
		log.Println("Create Group ", n)
		if _, err := stmt.Exec(s, rid); err != nil {
			log.Fatal(err)
		}
		lu, err := convertToMxid(ldap.LdapUsers(s))
		if err != nil {
			log.Fatal(err)
		}
		matrix.HandleCreatedUsers(rid, lu)
		if conf.Matrix.E2eEncryption == true {
			matrix.EnableEncryption(rid)
		}
	}
}

func handleCurrentGroups(lg []string) {
	stmt, err := db.Prepare("SELECT matrix_room FROM room_group_map WHERE ldap_group = $1 LIMIT 1")
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
		lu, err := convertToMxid(ldap.LdapUsers(s))
		if err != nil {
			log.Fatal(err)
		}
		mu := matrix.MatrixUsers(rid)

		matrix.HandleCreatedUsers(rid, Diff(mu, lu))
		matrix.HandleRemovedUsers(rid, Diff(lu, mu))
		matrix.SetRoomName(rid, ldap.LdapGroupName(s))
	}
}

func convertToMxid(lu []string, err_in error) (mu []string, err error) {
	err = err_in
	for _, e := range lu {
		mu = append(mu, "@"+e+":"+conf.Matrix.Homeserver)
	}
	return
}
