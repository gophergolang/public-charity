DROP TABLE IF EXISTS `surplus`;
--> statement-breakpoint
ALTER TABLE `users` DROP COLUMN `onboarding_step`;
--> statement-breakpoint
ALTER TABLE `users` DROP COLUMN `contact_prefs`;
--> statement-breakpoint
ALTER TABLE `offers` DROP COLUMN `available`;
--> statement-breakpoint
ALTER TABLE `availability` DROP COLUMN `note`;
