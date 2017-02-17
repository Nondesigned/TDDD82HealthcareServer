# TDDD82HealthcareServer
##How to run the server
Inorder to run the server you must have a self-signed TLS-certificate (cert.pem and key.unencrypted.pem). This can be generated with `openssl`. 
Or you can use the two test ones found in the drive. The private key must be unencrypted (the one in the drive is).

The server is started with the following commands:
```
go build
./TDDD82HealthcareServer
```

##Database cridentials
To use the database you must configure a file called `config.json` and
the format is supposed to be:
```
{
    "DBUser":"username",
    "DBPass":"password",
    "DBName":"itkand_2017_3_1 or itkand_2017_3_2"
}
```
##How to get a hashed password for testing
To get a hashed password for your test user please use `/create` to get a bcrypt-hash of 'kaffekaka'