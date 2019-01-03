package test

import (
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
	"github.com/jmoiron/sqlx"
)

func updateMaster(db *sqlx.DB) error {
	//change clients
	var sql = "UPDATE clients c SET c.deleted = MOD(c.deleted + 1, 2) WHERE c.version != 0 LIMIT 33"
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	//change programs
	sql = "UPDATE programs p SET p.version=0 LIMIT 1"
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	//change program_cards
	sql = "UPDATE program_cards SET version=0 LIMIT 1"
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	//change client_balance
	sql = "UPDATE client_balance SET version=0 LIMIT 1"
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func updateSlave(db *sqlx.DB) error {
	var sql = "UPDATE client_activity ca SET ca.version = 0 LIMIT 33"
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func loadActivity(db *sqlx.DB, source, docID string) (service.Activity, error) {
	var a service.Activity
	var sql = "SELECT source, doc_id, card, DATE_FORMAT(doc_date,'%Y-%m-%d %H:%i:%s') doc_date, doc_sum, bonuce_sum FROM client_activity WHERE source=? AND  doc_id=?"
	err := db.Get(&a, sql, source, docID)
	return a, err
}
