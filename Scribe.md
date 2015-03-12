# Introduction #
The Scribe is a REST+JSON service that provides persistence for jobs submitted to the Golem master.  This optional service integrates with the Golem stack using a SOA design.

# Details #
The service stack for the **Scribe/Master/Worker** receives authenticated client requests through the **Scribe**, which persists jobs to the database.  An asynchronous thread will submit the **pending** jobs from the database to the **Master**.  An asynchronous thread polls the **Master** for **running** jobs and update the database.  A **REST API** and user interface will be provided to monitor, **stop** and **delete** jobs.