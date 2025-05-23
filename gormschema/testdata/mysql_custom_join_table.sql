-- atlas:pos addresses[type=table] /internal/testdata/customjointable/models.go:17
-- atlas:pos people[type=table] /internal/testdata/customjointable/models.go:11
-- atlas:pos person_addresses[type=table] /internal/testdata/customjointable/models.go:22
-- atlas:pos top_crowded_addresses[type=view] /internal/testdata/customjointable/models.go:29

CREATE TABLE `addresses` (`id` bigint AUTO_INCREMENT,`name` longtext,PRIMARY KEY (`id`));
CREATE TABLE `people` (`id` bigint AUTO_INCREMENT,`name` longtext,PRIMARY KEY (`id`));
CREATE TABLE `person_addresses` (`person_id` bigint,`address_id` bigint,`created_at` datetime(3) NULL,`deleted_at` datetime(3) NULL,PRIMARY KEY (`person_id`,`address_id`));
CREATE VIEW top_crowded_addresses AS SELECT address_id, COUNT(person_id) AS count FROM person_addresses GROUP BY address_id ORDER BY count DESC LIMIT 10;
ALTER TABLE `person_addresses` ADD CONSTRAINT `fk_person_addresses_address` FOREIGN KEY (`address_id`) REFERENCES `addresses`(`id`);
ALTER TABLE `person_addresses` ADD CONSTRAINT `fk_person_addresses_person` FOREIGN KEY (`person_id`) REFERENCES `people`(`id`);
