CREATE EXTENSION postgis;

CREATE TABLE BUILDINGS (
	ID SERIAL PRIMARY KEY,
	DOITT_ID integer UNIQUE NOT NULL,
	YEAR integer,
	LASTMOD TIMESTAMPTZ NOT NULL,
	ROOF_HEIGHT real
);

/* Building Footprints uses NAD83 spatial reference system (a.k.a. 4269) */
SELECT AddGeometryColumn ('buildings','coords',4269,'POINT',2);