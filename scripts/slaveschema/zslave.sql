--
-- Скрипт сгенерирован Devart dbForge Studio for MySQL, Версия 7.3.131.0
-- Домашняя страница продукта: http://www.devart.com/ru/dbforge/mysql/studio
-- Дата скрипта: 28.11.2018 16:59:23
-- Версия сервера: 5.1.73-community
-- Версия клиента: 4.1
--


-- 
-- Отключение внешних ключей
-- 
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;

-- 
-- Установить режим SQL (SQL mode)
-- 
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;

-- 
-- Установка кодировки, с использованием которой клиент будет посылать запросы на сервер
--
SET NAMES 'utf8';

--
-- Установка базы данных по умолчанию
--
USE zslave;

--
-- Удалить таблицу "programs"
--
DROP TABLE IF EXISTS programs;

--
-- Удалить таблицу "program_cards"
--
DROP TABLE IF EXISTS program_cards;

--
-- Удалить таблицу "cnv_version"
--
DROP TABLE IF EXISTS cnv_version;

--
-- Удалить таблицу "cnv_source"
--
DROP TABLE IF EXISTS cnv_source;

--
-- Удалить таблицу "clients"
--
DROP TABLE IF EXISTS clients;

--
-- Удалить таблицу "client_state_msg"
--
DROP TABLE IF EXISTS client_state_msg;

--
-- Удалить таблицу "client_state"
--
DROP TABLE IF EXISTS client_state;

--
-- Удалить таблицу "client_balance"
--
DROP TABLE IF EXISTS client_balance;

--
-- Удалить таблицу "client_activity"
--
DROP TABLE IF EXISTS client_activity;

--
-- Установка базы данных по умолчанию
--
USE zslave;

--
-- Создать таблицу "client_activity"
--
CREATE TABLE client_activity (
  source char(2) NOT NULL,
  doc_id varchar(50) NOT NULL,
  card varchar(50) NOT NULL,
  doc_date datetime NOT NULL,
  doc_sum decimal(16, 2) NOT NULL DEFAULT 0.00,
  bonuce_sum decimal(16, 2) NOT NULL DEFAULT 0.00,
  version int(11) NOT NULL DEFAULT 0,
  deleted tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (source, doc_id),
  INDEX IDX_client_activity (source, version),
  INDEX IDX_client_activity_card (card)
)
ENGINE = INNODB
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "client_balance"
--
CREATE TABLE client_balance (
  card varchar(50) NOT NULL,
  balance_date date NOT NULL COMMENT 'the last day of the corresponding month',
  doc_sum decimal(16, 2) DEFAULT 0.00,
  bonuce_sum decimal(16, 2) DEFAULT 0.00,
  level int(5) NOT NULL DEFAULT 0,
  version int(11) NOT NULL DEFAULT 0,
  deleted tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (card, balance_date),
  INDEX IDX_client_balance_version (version)
)
ENGINE = INNODB
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "client_state"
--
CREATE TABLE client_state (
  id int(5) NOT NULL,
  name varchar(50) DEFAULT NULL,
  PRIMARY KEY (id)
)
ENGINE = INNODB
AVG_ROW_LENGTH = 1638
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "client_state_msg"
--
CREATE TABLE client_state_msg (
  id int(5) NOT NULL,
  web_comment varchar(500) DEFAULT NULL,
  PRIMARY KEY (id),
  CONSTRAINT FK_client_state_msg_id FOREIGN KEY (id)
  REFERENCES client_state (id) ON DELETE CASCADE ON UPDATE CASCADE
)
ENGINE = INNODB
AVG_ROW_LENGTH = 2340
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "clients"
--
CREATE TABLE clients (
  program int(5) NOT NULL,
  card varchar(50) NOT NULL,
  state int(5) NOT NULL DEFAULT 1,
  birthday datetime DEFAULT NULL,
  version int(11) NOT NULL DEFAULT 0,
  deleted tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (program, card),
  INDEX IDX_clients_version (version),
  CONSTRAINT FK_clients_program FOREIGN KEY (program)
  REFERENCES programs (id) ON DELETE NO ACTION ON UPDATE RESTRICT
)
ENGINE = INNODB
AVG_ROW_LENGTH = 141
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "cnv_source"
--
CREATE TABLE cnv_source (
  id char(2) NOT NULL,
  name varchar(50) DEFAULT NULL,
  url varchar(100) DEFAULT NULL,
  PRIMARY KEY (id)
)
ENGINE = INNODB
AVG_ROW_LENGTH = 16384
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "cnv_version"
--
CREATE TABLE cnv_version (
  source char(2) NOT NULL DEFAULT '',
  table_name varchar(40) NOT NULL,
  latest_version int(11) NOT NULL DEFAULT 0,
  syncorder int(5) NOT NULL DEFAULT 0,
  PRIMARY KEY (source, table_name)
)
ENGINE = INNODB
AVG_ROW_LENGTH = 3276
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "program_cards"
--
CREATE TABLE program_cards (
  id int(11) NOT NULL AUTO_INCREMENT,
  program int(5) NOT NULL,
  card_start varchar(50) NOT NULL,
  card_end varchar(50) NOT NULL,
  active tinyint(1) DEFAULT 1,
  card_len int(11) NOT NULL,
  check_issued tinyint(1) DEFAULT 0,
  version int(11) NOT NULL DEFAULT 0,
  deleted tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  INDEX IDX_program_cards (card_len, card_start, card_end),
  INDEX IDX_program_cards_version (version),
  CONSTRAINT FK_program_cards_program FOREIGN KEY (program)
  REFERENCES programs (id) ON DELETE CASCADE ON UPDATE CASCADE
)
ENGINE = INNODB
AVG_ROW_LENGTH = 2340
CHARACTER SET utf8
COLLATE utf8_general_ci;

--
-- Создать таблицу "programs"
--
CREATE TABLE programs (
  id int(5) NOT NULL AUTO_INCREMENT,
  name varchar(50) DEFAULT NULL,
  alias varchar(50) DEFAULT NULL,
  external tinyint(1) DEFAULT 0,
  active tinyint(1) DEFAULT 1,
  version int(11) NOT NULL DEFAULT 0,
  deleted tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  INDEX IDX_programs_version (version)
)
ENGINE = INNODB
AVG_ROW_LENGTH = 16384
CHARACTER SET utf8
COLLATE utf8_general_ci;

-- 
-- Вывод данных для таблицы client_activity
--
-- Таблица zslave.client_activity не содержит данных

-- 
-- Вывод данных для таблицы client_balance
--
-- Таблица zslave.client_balance не содержит данных

-- 
-- Вывод данных для таблицы client_state
--
INSERT INTO client_state VALUES
(-1001, 'Соглашение'),
(-1000, 'Инициализация'),
(-12, 'Карта не выдана'),
(-11, 'Не верный статус'),
(-10, 'Указана не верная карта'),
(-1, 'Ошибка базы данных'),
(1, 'Выдана'),
(5, 'Регистрация'),
(10, 'Уточнение анкеты'),
(100, 'Зарегистрирован');

-- 
-- Вывод данных для таблицы client_state_msg
--
INSERT INTO client_state_msg VALUES
(-1001, 'С <a href="http://www.tut.by" target="_blank" >правилами и условиями</a>  пользования дисконтной картой ознакомлен(а) и согласен(а). Приложение №1.'),
(-1000, 'Укажите код указанный на карте'),
(-12, 'Нет данных о выдаче карты пользователю. Повторите попытку позже. '),
(-10, ' <span color="blue"> Проверьте код карты</span>'),
(-1, 'Сервис не доступен. Попробуйте повторить попытку позже.'),
(5, 'Ваша анкета ожидает поверки на корректность заполнения.'),
(10, 'Ваша анкета не корректно заполнена. Для уточнения анкетных данных обратитесь в место получения карты.'),
(100, 'Ваша анкета зарегистрирована.');

-- 
-- Вывод данных для таблицы clients
--
-- Таблица zslave.clients не содержит данных

-- 
-- Вывод данных для таблицы cnv_source
--
INSERT INTO cnv_source VALUES
('00', 'Центральная база', NULL);

-- 
-- Вывод данных для таблицы cnv_version
--
INSERT INTO cnv_version VALUES
('', 'client_activity', 0, 0),
('00', 'clients', 2, 0),
('00', 'client_balance', 3, 0),
('00', 'programs', 0, 0),
('00', 'program_cards', 1, 0);

-- 
-- Вывод данных для таблицы program_cards
--
-- Таблица zslave.program_cards не содержит данных

-- 
-- Вывод данных для таблицы programs
--
-- Таблица zslave.programs не содержит данных

--
-- Установка базы данных по умолчанию
--
USE zslave;

DELIMITER $$

--
-- Удалить триггер "client_activity_tg_ai"
--
DROP TRIGGER IF EXISTS client_activity_tg_ai$$

DELIMITER ;

--
-- Установка базы данных по умолчанию
--
USE zslave;

DELIMITER $$

--
-- Создать триггер "client_activity_tg_ai"
--
CREATE
DEFINER = 'root'@'localhost'
TRIGGER client_activity_tg_ai
AFTER INSERT
ON client_activity
FOR EACH ROW
BEGIN
  DECLARE vdate date;
  SET vdate = LAST_DAY(NEW.doc_date);
  IF vdate = LAST_DAY(CURDATE())
  THEN
    INSERT INTO client_balance (card, balance_date, doc_sum, bonuce_sum)
      VALUES (NEW.card, vdate, NEW.doc_sum, NEW.bonuce_sum)
    ON DUPLICATE KEY UPDATE doc_sum = doc_sum + NEW.doc_sum, bonuce_sum = bonuce_sum + NEW.bonuce_sum;
  END IF;
END
$$

DELIMITER ;
-- 
-- Восстановить предыдущий режим SQL (SQL mode)
-- 
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;

-- 
-- Включение внешних ключей
-- 
/*!40014 SET FOREIGN_KEY_CHECKS = @OLD_FOREIGN_KEY_CHECKS */;
