CREATE TABLE tenant.parents (
	tenant_id INT,
	id INT,
	name VARCHAR(255),
	PRIMARY KEY (tenant_id, id)
);

CREATE TABLE tenant.children (
	id INT PRIMARY KEY,
	tenant_id INT,
	parent_id INT,
	FOREIGN KEY (tenant_id, parent_id) REFERENCES tenant.parents(tenant_id, id) ON DELETE CASCADE
);
