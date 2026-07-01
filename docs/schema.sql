
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
SET @MYSQLDUMP_TEMP_LOG_BIN = @@SESSION.SQL_LOG_BIN;
SET @@SESSION.SQL_LOG_BIN= 0;
SET @@GLOBAL.GTID_PURGED=/*!80000 '+'*/ '40b99432-5687-11f1-aa0d-9a8df31aa9fe:1-35';
DROP TABLE IF EXISTS `checkin_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `checkin_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `invitation_id` bigint unsigned NOT NULL,
  `event_type` varchar(30) NOT NULL,
  `checked_in_pax` smallint unsigned NOT NULL,
  `checked_in_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `scanner_name` varchar(150) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_invitation_event` (`invitation_id`,`event_type`),
  CONSTRAINT `fk_invitations_checkin_logs` FOREIGN KEY (`invitation_id`) REFERENCES `invitations` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tag` varchar(100) NOT NULL,
  `couple_name` varchar(150) NOT NULL,
  `holy_matrimony_date` datetime DEFAULT NULL,
  `holy_matrimony_location` text,
  `reception_date` datetime DEFAULT NULL,
  `reception_location` text,
  `gift_address` text,
  `bank_account` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_events_tag` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `invitations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `invitations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `event_id` bigint unsigned NOT NULL,
  `guest_name` varchar(150) NOT NULL,
  `whatsapp_number` varchar(30) NOT NULL,
  `invitation_code` varchar(100) NOT NULL,
  `invitation_status` varchar(30) NOT NULL DEFAULT 'imported',
  `rsvp_status` varchar(30) NOT NULL DEFAULT 'not_answered',
  `pax_count` smallint unsigned DEFAULT NULL,
  `event_choice` varchar(30) DEFAULT NULL,
  `gift_interest` varchar(30) NOT NULL DEFAULT 'not_asked',
  `qr_code_token` varchar(150) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_event_whatsapp` (`event_id`,`whatsapp_number`),
  UNIQUE KEY `idx_invitations_invitation_code` (`invitation_code`),
  UNIQUE KEY `idx_invitations_qr_code_token` (`qr_code_token`),
  KEY `idx_invitations_event_id` (`event_id`),
  KEY `idx_invitations_invitation_status` (`invitation_status`),
  KEY `idx_invitations_rsvp_status` (`rsvp_status`),
  KEY `idx_invitations_event_choice` (`event_choice`),
  CONSTRAINT `fk_events_invitations` FOREIGN KEY (`event_id`) REFERENCES `events` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
DROP TABLE IF EXISTS `whatsapp_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `whatsapp_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `invitation_id` bigint unsigned NOT NULL,
  `message_type` varchar(40) DEFAULT NULL,
  `direction` varchar(20) DEFAULT NULL,
  `whatsapp_message_id` varchar(255) DEFAULT NULL,
  `status` varchar(20) NOT NULL DEFAULT 'pending',
  `payload` json DEFAULT NULL,
  `meta_response` json DEFAULT NULL,
  `meta_error` json DEFAULT NULL,
  `sent_at` datetime DEFAULT NULL,
  `received_at` datetime DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_whatsapp_logs_invitation_id` (`invitation_id`),
  KEY `idx_whatsapp_logs_message_type` (`message_type`),
  KEY `idx_whatsapp_logs_whatsapp_message_id` (`whatsapp_message_id`),
  KEY `idx_whatsapp_logs_status` (`status`),
  CONSTRAINT `fk_invitations_whatsapp_logs` FOREIGN KEY (`invitation_id`) REFERENCES `invitations` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
SET @@SESSION.SQL_LOG_BIN = @MYSQLDUMP_TEMP_LOG_BIN;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

