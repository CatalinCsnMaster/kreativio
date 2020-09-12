# Kreativio deployment

This repository has been set up to aid the migration of Kreativio.ro.
The front-end is custom build in Angular and packaged in a static webserver.
The back end is a collection of APIs which rely on gRPC-web for communiction with the front-end.

The configuration in this repository can be considered as an example, but can be deployed in production.
When using this example in production, it should be taken in consideration that the databse is a single instance and no back-ups are configured.

## Requirements

### Server Host

This repository assumes Docker and docker-compose to be available on a server. Furthermore, DNS should be set up correctly to point to this server.
At least the following domains should point to the IP address of the server:

- `kreativio.ro`
- `auth.kreativio.ro`
- `admin.auth.kreativio.ro`

### E-mail

An authenticated SMTP server has to be available for outgoing mail from the APIs, such as contact form and order confirmations.

### S3

Media and images are stored on S3 compatible storage. An API key and Secret is required.

## Starting the project

Follow the following steps to start the project.

### TLS certificates

The project uses Treafik as reverse proxy for HTTP routing and TLS termination.
Traefik will obtain the required certificates from LetsEncrypt automatically.
First, the certificate store needs to be created:

````
mkdir data/traefik
touch data/traefik/acme.json
chmod 600 data/traefik/acme.json
````

IMPORTANT: DNS has to be configured correctly before starting the project, or else certificate generation will fail!

### 1: Configuration

The project is pre-configured to use a PostgreSQL Docker image.
If an external database is used instead, it needs be re-configured for all services.

#### SMTP e-mail

The e-mail server credentials are at the moment configured to use Mohlmann Solutions' servers.
They need to be reconfigured in the files:

````
config/authenticator/server.json
config/shop/shop.json
````

#### S3

S3 configuration is emptied at the moment, as the existing keys are private to Mohlmann Solutions. Please add the correct config to:

````
config/shop/imageapi.env
````

### 2: Start DB server

The first time, PostgreSQL needs to be started before anything else.
It needs some time to run InitDB, during which other services would fail.

Run:

````
docker-compose up -d && docker-compose logs -f
````

Once you see a log entry like:

````
LOG:  database system is ready to accept connections
````

Hit `CTRL+C` to exit the logger.

### 3: Bring up the rest of the project

````
docker-compose up -d
docker-compose logs -f
````

This should bring all the services up, create database tables (migrations) and create the first user.

The first user is defined in `config/authenticator/server.json` as:

````
"bootsrap": [
    {
      "Email": "kreativio@yahoo.com",
      "Name": "Viorel Ghita",
      "Password": "NiZ0Ut0eijeefi9k",
      ....
````

### 4: Secure the user

After a minute or so, Traefik should have obtained the TLS certificates.
You can browse to https://admin.auth.kreativio.ro
Use the above credentials to log in.
Once logged in, new users can be created.
The reset password functionality can be used to create a new password.

### 5: Check the website

Browse to https://kreativio.ro. It should show a website with some content.
A Database error will be shown, as there are no articles in the DB yet.

Browse to https://kreativio.ro/admin to login and start adding articles.
