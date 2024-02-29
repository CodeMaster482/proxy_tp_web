
CREATE TABLE IF NOT EXISTS request (
	id 		   		SERIAL 		  PRIMARY KEY			   	  NOT NULL,
	method 	   		TEXT 		  CHECK(length(method) < 10)  NOT NULL,
	"host" 	   		TEXT 		  CHECK(length("host") < 500)  NOT NULL,
	"path"			TEXT 		  CHECK(length("path") < 500)  NOT NULL,
	headers    		JSONB,
	query_params	JSONB,
	post_params		JSONB,
	cookies			JSONB,
	body 	   		TEXT,
	created_at 		TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP   NOT NULL
);

CREATE TABLE IF NOT EXISTS response (
	id 				SERIAL 		PRIMARY KEY 				NOT NULL,
	request_id		INTEGER									NOT NULL,
	status_code 	INTEGER									NOT NULL,
	headers 		JSONB,
	body 			TEXT,
	created_at 		TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP   NOT NULL,
	FOREIGN KEY (request_id) REFERENCES request(id) ON DELETE CASCADE
);