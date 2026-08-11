/*M!999999\- enable the sandbox mode */ 
-- MariaDB dump 10.19-11.8.6-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: snippetbox
-- ------------------------------------------------------
-- Server version	11.8.6-MariaDB-5ubuntu0.1 from Ubuntu

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*M!100616 SET @OLD_NOTE_VERBOSITY=@@NOTE_VERBOSITY, NOTE_VERBOSITY=0 */;

--
-- Current Database: `snippetbox`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `snippetbox` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */;

USE `snippetbox`;

--
-- Table structure for table `sessions`
--

DROP TABLE IF EXISTS `sessions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `sessions` (
  `token` char(43) NOT NULL,
  `data` blob NOT NULL,
  `expiry` timestamp(6) NOT NULL,
  PRIMARY KEY (`token`),
  KEY `sessions_expiry_idx` (`expiry`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `sessions`
--

SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT, @@AUTOCOMMIT=0;
LOCK TABLES `sessions` WRITE;
/*!40000 ALTER TABLE `sessions` DISABLE KEYS */;
INSERT INTO `sessions` VALUES
('-o8hJslAVJo7_jbrb0LfKD198V5POSAdS_uqp64gNvI','%ÿ€\0Deadlineÿ‚\0Valuesÿ„\0\0\0ÿTimeÿ‚\0\0\0\'ÿƒmap[string]interface {}ÿ„\0\0\0ÿ€\0\0\0âÑ\Z\"ÜCÓÿÿ\0\0','2026-08-18 10:18:02.584860');
/*!40000 ALTER TABLE `sessions` ENABLE KEYS */;
UNLOCK TABLES;
COMMIT;
SET AUTOCOMMIT=@OLD_AUTOCOMMIT;

--
-- Table structure for table `snippets`
--

DROP TABLE IF EXISTS `snippets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `snippets` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(100) NOT NULL,
  `content` text NOT NULL,
  `created` datetime NOT NULL,
  `expires` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_snippets_created` (`created`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `snippets`
--

SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT, @@AUTOCOMMIT=0;
LOCK TABLES `snippets` WRITE;
/*!40000 ALTER TABLE `snippets` DISABLE KEYS */;
INSERT INTO `snippets` VALUES
(1,'An old silent pond','An old silent pond...\nA frog jumps into the pond,\nsplash! Silence again.\n\nâ€“ Matsuo BashÅ','2026-08-06 00:59:23','2027-08-06 00:59:23'),
(2,'Over the wintry forest','Over the wintry\nforest, winds howl in rage\nwith no leaves to blow.\n\nâ€“ Natsume Soseki','2026-08-06 00:59:23','2027-08-06 00:59:23'),
(3,'First autumn morning','First autumn morning\nthe mirror I stare into\nshows my father\'s face.\n\nâ€“ Murakami Kijo','2026-08-06 00:59:24','2026-08-13 00:59:24'),
(4,'O snail','O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\nâ€“ Kobayashi Issa','2026-08-10 03:01:24','2026-08-17 03:01:24'),
(5,'From time to time','From time to time\r\nThe clouds give rest\r\nTo the moon-beholders.\r\n\r\n- Matsumo Basho','2026-08-10 03:08:56','2026-08-17 03:08:56'),
(6,'','','2026-08-10 03:12:00','2027-08-10 03:12:00'),
(7,'sudo mariadb -D snippetbox','[sudo: authenticate] Password:           \r\nReading table information for completion of table and column names\r\nYou can turn off this feature to get a quicker startup with -A\r\n\r\nWelcome to the MariaDB monitor.  Commands end with ; or \\g.\r\nYour MariaDB connection id is 2392670\r\nServer version: 11.8.6-MariaDB-5ubuntu0.1 from Ubuntu -- Please help get to 10k stars at https://github.com/MariaDB/Server\r\n\r\nCopyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.\r\n\r\nType \'help;\' or \'\\h\' for help. Type \'\\c\' to clear the current input statement.\r\n\r\nMariaDB [snippetbox]> USE snippetbox;\r\nDatabase changed\r\nMariaDB [snippetbox]> CREATE TABLE sessions (\r\n    ->     token CHAR(43) PRIMARY KEY,\r\n    ->     data BLOB NOT NULL,\r\n    ->     expiry TIMESTAMP(6) NOT NULL\r\n    -> );\r\nQuery OK, 0 rows affected (0.135 sec)\r\n\r\nMariaDB [snippetbox]> CREATE INDEX sessions_expiry_idx ON sessions (expiry);\r\nQuery OK, 0 rows affected (0.258 sec)\r\nRecords: 0  Duplicates: 0  Warnings: 0\r\n\r\nMariaDB [snippetbox]> ^DBye\r\n','2026-08-10 22:18:02','2027-08-10 22:18:02'),
(8,'amend','[main fedc216] Automatic form parsing\r\n Date: Tue Aug 11 09:21:57 2026 +1200\r\n 6 files changed, 43 insertions(+), 34 deletions(-)','2026-08-10 22:28:17','2027-08-10 22:28:17'),
(9,'go get github.com/alexedwards/scs/mysqlstore@latest','go: downloading github.com/alexedwards/scs v1.4.1\r\ngo: downloading github.com/alexedwards/scs/mysqlstore v0.0.0-20251002162104-209de6e426de\r\ngo: added github.com/alexedwards/scs/mysqlstore v0.0.0-20251002162104-209de6e426de','2026-08-10 22:30:37','2027-08-10 22:30:37');
/*!40000 ALTER TABLE `snippets` ENABLE KEYS */;
UNLOCK TABLES;
COMMIT;
SET AUTOCOMMIT=@OLD_AUTOCOMMIT;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `hashed_password` char(60) NOT NULL,
  `created` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `users_uc_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT, @@AUTOCOMMIT=0;
LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
COMMIT;
SET AUTOCOMMIT=@OLD_AUTOCOMMIT;

--
-- Dumping events for database 'snippetbox'
--

--
-- Dumping routines for database 'snippetbox'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*M!100616 SET NOTE_VERBOSITY=@OLD_NOTE_VERBOSITY */;

-- Dump completed on 2026-08-11 13:49:37
