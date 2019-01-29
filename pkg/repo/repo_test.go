package repo

import (
	"context"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

func newDb(cnn, folder string) (service.Repository, *sqlx.DB, error) {
	//"root:3411@tcp(127.0.0.1:3306)/pshdata"
	var db *sqlx.DB
	db, err := sqlx.Connect("mysql", cnn)
	if err != nil {
		return nil, nil, err
	}
	return New(db, folder), db, nil
}

/*
func TestFixVersion(t *testing.T){
	mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
}
*/

func TestFixVersionMaster(t *testing.T) {
	//var mdb service.Repository
	mrep, mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()

	ver0, err := mrep.ListVersion(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

	//change clients
	var sql = "UPDATE clients c SET c.deleted = MOD(c.deleted + 1, 2) WHERE c.version != 0 LIMIT 33"
	_, err = mdb.Exec(sql)
	if err != nil {
		t.Error(err)
	}

	//change programs
	sql = "UPDATE programs p SET p.version=0 LIMIT 1"
	_, err = mdb.Exec(sql)
	if err != nil {
		t.Error(err)
	}

	//change program_cards
	sql = "UPDATE program_cards SET version=0 LIMIT 1"
	_, err = mdb.Exec(sql)
	if err != nil {
		t.Error(err)
	}

	//change client_balance
	sql = "UPDATE client_balance SET version=0 LIMIT 1"
	_, err = mdb.Exec(sql)
	if err != nil {
		t.Error(err)
	}

	err = mrep.FixVersions(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}

	ver, err := mrep.ListVersion(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver)

	//check if version updated
	for _, v0 := range ver0 {
		for _, v := range ver {
			if v0.Table == v.Table && (v0.Version+1) != v.Version {
				t.Error(v0.Table, " Expected version ", (v0.Version + 1), ", got ", v.Version)
			}
		}
	}

}

func TestSyncSlave(t *testing.T) {
	mrep, mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	srep, sdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/zslave", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer sdb.Close()

	//get table versions in slave
	ver0, err := srep.ListVersion(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

	//generate sync packs in master
	packList := make([]service.VersionPack, 0, len(ver0))
	for _, v0 := range ver0 {
		var fileName = "00_" + v0.Table + "_" + strconv.FormatInt(int64(v0.Version), 10) + ".dat"
		p, err := mrep.CreatePack(context.Background(), "00", v0.Table, fileName, v0.Version)
		if err != nil {
			t.Error(err)
		}
		packList = append(packList, p)
	}
	t.Log(packList)

	//apply packs in slave
	for _, p := range packList {
		if p.Pack == "" {
			t.Log("Version not changed or error in CreatePack", p)
		}
		err = srep.ExecPack(context.Background(), p)
		if err != nil {
			t.Error(err)
		}
	}

	//check versions vs master
	ver0, err = srep.ListVersion(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

	ver, err := mrep.ListVersion(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver)
	for _, v0 := range ver0 {
		var found = false
		for _, v := range ver {
			if v0.Table == v.Table {
				found = true
				if v0.Version != v.Version {
					t.Error(v0.Table, "Version not changed. Slave version ", v0.Version, ". Master version ", v.Version)
				}
			}
		}
		if !found {
			t.Error("Не найдена таблица ", v0.Table)
		}
	}

}

func TestFixVersionSlave(t *testing.T) {
	//var mdb service.Repository
	mrep, mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/zslave", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()

	ver0, err := mrep.ListVersion(context.Background(), "zs")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

	//change client_activity
	var sql = "UPDATE client_activity ca SET ca.version = 0 LIMIT 33"
	_, err = mdb.Exec(sql)
	if err != nil {
		t.Error(err)
	}

	//client
	err = mrep.FixVersions(context.Background(), "zs")
	if err != nil {
		t.Error(err)
	}

	ver, err := mrep.ListVersion(context.Background(), "zs")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver)

	//check if version updated
	for _, v0 := range ver0 {
		for _, v := range ver {
			if v0.Table == v.Table && (v0.Version+1) != v.Version {
				t.Error(v0.Table, " Expected version ", (v0.Version + 1), ", got ", v.Version)
			}
		}
	}

}

func TestSyncMaster(t *testing.T) {
	mrep, mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	srep, sdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/zslave", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer sdb.Close()

	//get table versions in  master
	ver0, err := mrep.ListVersion(context.Background(), "zs")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

	//generate sync packs in slave
	packList := make([]service.VersionPack, 0, len(ver0))
	for _, v0 := range ver0 {
		var fileName = "zs_" + v0.Table + "_" + strconv.FormatInt(int64(v0.Version), 10) + ".dat"
		p, err := srep.CreatePack(context.Background(), "zs", v0.Table, fileName, v0.Version)
		if err != nil {
			t.Error(err)
		}
		packList = append(packList, p)
	}
	t.Log(packList)

	//apply packs in master
	for _, p := range packList {
		if p.Pack == "" {
			t.Log("Version not changed or error in CreatePack", p)
		}
		err = mrep.ExecPack(context.Background(), p)
		if err != nil {
			t.Error(err)
		}
	}

	//check versions vs slave
	ver0, err = mrep.ListVersion(context.Background(), "zs")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

	ver, err := srep.ListVersion(context.Background(), "zs")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver)
	for _, v0 := range ver0 {
		var found = false
		for _, v := range ver {
			if v0.Table == v.Table {
				found = true
				if v0.Version != v.Version {
					t.Error(v0.Table, "Version not changed. Master version ", v0.Version, ". Slave version ", v.Version)
				}
			}
		}
		if !found {
			t.Error("Не найдена таблица ", v0.Table)
		}
	}

}

func TestCalcs(t *testing.T) {
	mrep, mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()

	//get first day of month
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentMonth = currentMonth - 1
	if currentMonth < 1 {
		currentYear = currentYear - 1
		currentMonth = 12
	}
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	/*
		fmt.Println(firstOfMonth)
		var sql = "select timediff(now(),convert_tz(now(),@@session.time_zone,'+00:00'));"
		var str string
		mdb.Get(&str, sql)
		fmt.Println(str)

		err = mdb.Get(&str, "CALL recalc_level(?)", firstOfMonth.Format("2006-01-02"))
		if err != nil {
			t.Error(err)
		}
		fmt.Println(str)
	*/

	//clear current
	var sql = "UPDATE client_balance SET bonuce_sum = 0, level = 0 WHERE balance_date = LAST_DAY(CURDATE())"
	_, err = mdb.Exec(sql)
	if err != nil {
		t.Error(err)
	}

	//calc balance
	err = mrep.CalcBalance(context.Background(), firstOfMonth)
	if err != nil {
		t.Error(err)
	}

	//calc levels
	err = mrep.CalcLevels(context.Background(), firstOfMonth)
	if err != nil {
		t.Error(err)
	}

	//check getlevel
	sql = "SELECT c.card, cb.level FROM programs p" +
		" INNER JOIN clients c ON p.id = c.program" +
		" INNER JOIN client_balance cb ON c.card = cb.card AND cb.balance_date = ADDDATE(CURDATE(), -DAY(CURDATE()))" +
		" WHERE p.external = 0 AND c.state >= 5 AND cb.level > 0" +
		" LIMIT 1"
	data := struct {
		Card  string
		Level int
	}{
		"",
		0,
	}
	err = mdb.Get(&data, sql)
	if err != nil {
		t.Error(err)
	}

	lvl, err := mrep.GetLevel(context.Background(), data.Card)
	if err != nil {
		t.Error(err)
	}
	if lvl != data.Level {
		t.Error("Wrong card level, expected ", data.Level, ", got ", lvl, ". card ", data.Card)
	}
}

func TestListSource(t *testing.T) {
	mrep, mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()

	//count using direct sql
	var cnt int
	var sql = "SELECT count(*) res FROM cnv_source cs WHERE cs.id != ?"
	err = mdb.Get(&cnt, sql, "00")
	if err != nil {
		t.Error(err)
	}

	lst, err := mrep.ListSource(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(lst)

	if len(lst) != cnt {
		t.Error("Wrong source count, expected ", cnt, ", got ", len(lst))
	}
}
