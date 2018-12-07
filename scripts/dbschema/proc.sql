DELIMITER $$

--
-- Создать процедуру "clients_getcnv"
--
CREATE 
PROCEDURE clients_getcnv (IN p_source char(2), IN p_table varchar(40), IN p_start int, IN p_file varchar(255))
BEGIN
  DECLARE v_maxVer int;
  -- DECLARE v_sql varchar(1000);

  SELECT cv.latest_version
  INTO v_maxVer
    FROM cnv_version cv
    WHERE cv.source = p_source
      AND cv.table_name = p_table;

  IF v_maxVer > p_start
  THEN
    /*
    SELECT program, card, state, birthday, version
    INTO OUTFILE 'C:/tmp/result1.text'
      FROM clients c
    WHERE c.version> p_start AND c.version<= v_maxVer;
  */

    SET @v_sql = CONCAT(
    'SELECT program, card, state, birthday, version',
    ' INTO OUTFILE ''', p_file, '''',
    ' FROM clients c',
    ' WHERE c.version > ', p_start, ' AND c.version <= ', v_maxVer);
    PREPARE st FROM @v_sql;
    EXECUTE st;
    DEALLOCATE PREPARE st;
  END IF;

  SELECT v_maxVer;

END
$$

DELIMITER ;