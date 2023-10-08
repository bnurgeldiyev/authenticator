# Authenticator

--------------

This is authentication microservice.

We use user authentication for:

* Admin panels. Create user or give permissions some users. Create contents, posts etc.
* Verify user permissions or is this user authorized?
* Verifying token

If each our projects we use authenticator, we are copying some files for run authentication.

I thought it would be useful to make this microservice.

Let's talk about our stack.

* Golang - programming language
* Postgresql - to save user data
* Redis - to save refresh token

I use for authentication JWT(https://jwt.io), protocol GRPC(https://grpc.io).

---

Our microservice has 5 api's.

### Create
* input
  * username
  * password

username must be unique, if this username already exists in our database we return error code 6.

### Auth
* input
  * username
  * password

If username and password correct
* output
  * access_token
  * refresh_token

### Delete
* input
  * username

If we have this username, there will be deleted or returned error code 5.

### ValidateToken
* input
  * access_token

If this token is correct returned error code 0 or 16.

### UpdateToken
(If our accessToken is expired)
* input
  * access_token
  * refresh_token

if access_token is expired and refresh_token is correct we are returning new access and refresh tokens.

___

## Run
* you can install golang, postgresql, redis manually
* change name .sample.env to .env and set your dependencies
* make run

### If your have docker installed
```
    make start
```
