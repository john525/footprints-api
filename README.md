<h1>Building Footprints Go API</h1>
A Go API for accessing and manipulating data from the NYC DOITT's "Building Footprints" dataset, a feature class which stores polygonal outlines of the buildings in New York City.

The ETL pipeline works by downloading a GeoJSON data file from NYC DOITT once every 24 hours and loading it into a PostgreSQL database which was configured with the extension PostGIS, used for storing geographic and geometric data. The pipeline checks every feature's last modified date to avoid overwriting new data with old data.

I used the geometry data types supported in PostGIS as the documentation recommends using geometric rather than geographic data types for data spanning small regions, such as cities, due to performance costs.

The application loads and provides access to the following data:
<ul>
<li>ID - the table's primary key</li>
<li>DOITT ID - a unique identifier assigned to each building in by the DOITT</li>
<li>Year - the year the building was constructed (or is scheduled to be constructed)</li>
<li>Last Modified - a timestamp stating when the NYC DOITT last updated data on this feature</li>
<li>Roof Height - the height that the roof extends above ground elevation.
<li>Coordinates - the location of the feature according to the NAD83 coordinate system (stored as a PostGIS geometric Point data type)
</ul>

<h2>Instructions</h2>
To set up the postgres server, install postgres and postgis. For OSX:
<code>brew install postgres && brew install postgis</code>

Then start the postgres server as follows:
<code>pg_ctl -D /usr/local/var/postgres start</code>

Create a database called "nyc_buildings", and connect to it as the user postgres.
The login credentials should be set to those specified in server/database.go.

After ensuring that user postgres owns nyc_buildings, run the script create_table.sql in nyc_buildings to create a table named BUILDINGS.

Then run "go build" to compile the code.

Find the binary "footprints-api" and run it from terminal:
<code>./footprints-api</code>
You may use the flag "-port" to specify which port the API can be accessed from. The default port is set to 15000.

When the executable is run, a thread begins which automatically downloads a new version of the Building Footprints GeoJSON file from NYC Open Data every 24 hours.
To see the server's current activity, type "status" into the CLI.
For more options, "help" can be entered.

Once the server is running on your local machine, the full data on a building whose DOITT ID is known can be retrieved from the postgres datastore in the following fashion:
http://localhost:15000/building?doitt_id=1205352

The average height of buildings constructed between two years can be found as follows:
http://localhost:15000/avg_height_between_years?min=1981&max=2016

The file with name "logfile" stores the server's history for debugging purposes.

For more information on Building Footprints, visit the description on the City of New York's github here: https://github.com/CityOfNewYork/nyc-geo-metadata/blob/master/Metadata/Metadata_BuildingFootprints.md
