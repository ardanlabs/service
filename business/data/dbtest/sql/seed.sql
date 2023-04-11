INSERT INTO products (product_id, user_id, name, cost, quantity, date_created, date_updated) VALUES
	('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'Comic Books', 50, 42, '2019-01-01 00:00:01.000001+00', '2019-01-01 00:00:01.000001+00'),
	('72f8b983-3eb4-48db-9ed0-e45cc6bd716b', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'McDonalds Toys', 75, 120, '2019-01-01 00:00:02.000001+00', '2019-01-01 00:00:02.000001+00')
	ON CONFLICT DO NOTHING;

INSERT INTO sales (sale_id, user_id, product_id, quantity, paid, date_created) VALUES
	('98b6d4b8-f04b-4c79-8c2e-a0aef46854b7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 2, 100, '2019-01-01 00:00:03.000001+00'),
	('85f6fb09-eb05-4874-ae39-82d1a30fe0d7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 5, 250, '2019-01-01 00:00:04.000001+00'),
	('a235be9e-ab5d-44e6-a987-fa1c749264c7', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '72f8b983-3eb4-48db-9ed0-e45cc6bd716b', 3, 225, '2019-01-01 00:00:05.000001+00')
	ON CONFLICT DO NOTHING;
