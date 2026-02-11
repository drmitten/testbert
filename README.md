# TestBert

A coding exercise for interview purposes.

### Building and Testing

- Create a `.env` file in the root folder of the project. (Take a look at `sample.env` to see the variables that must be supplied along with optional variables that can be used to override default values.)

- From that same folder, run `docker compose up -d` in a terminal.  This will stand up both a PostgreSQL server and the TestBert server in two docker containers.  (`docker compose down` to stop both servers)

- Still in the root folder of the project, run `go test server/test/* -v` to run the suite of automated tests against the service running in the docker container.

### Assumption

Users are authenticated via some other internal service and include a signed JWT with each request. (No JWT is necessary to access a shared collection via a token).  For simplicity, the signing key is specified via env variable.

### Improvements

- Add methods for users to be able to list what collections are available to them.

- Publish the access events to a queue for processing

### Comments

I chose to use PostgreSQL for storage because it's proven and fairly simple to work with, but I also implemented an in-memory datastore just to give an idea of how it might work using something like Redis.
