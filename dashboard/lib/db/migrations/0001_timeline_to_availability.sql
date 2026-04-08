DROP TABLE IF EXISTS `timeline`;
--> statement-breakpoint
CREATE TABLE `availability` (
	`user_id` text NOT NULL,
	`day` text NOT NULL,
	`slot` text NOT NULL,
	`note` text,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `availability_pk` ON `availability` (`user_id`,`day`,`slot`);
