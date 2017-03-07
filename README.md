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
```json
{
    "DBUser":"username",
    "DBPass":"password",
    "DBName":"@tcp(db-und.ida.liu.se:3306/itkand_2017_3_1 or @tcp(db-und.ida.liu.se:3306/itkand_2017_3_2"
}
```
##Format of JSON objects
To check the format of the JSON object you send to the server through /login and /contacts check the class `structures.go`
Example of input for /login:
```json
{
	"card":123,
	"password":"kaffekaka",
	"fcmtoken":"rndtokenstring"
}
```
##How to create a user
Currently a default user is created when performing a GET on someaddress:port/create. This should 

##Try login
After creating a user you can attempt to login with NFC-id:123, password:kaffekaka and fcmtoken:randomstring by performing a POST on someaddress:port/login

##Retrieve contactlist
By performing a GET on someaddress:port/contacs with a header containing the token (see documentation below for sending a token to the server).

##Tokens
This sections goes through how the tokens are implemented.

###Getting a token from the server
The token is returned when a user is successfully logged-in. See the following example:
```json
{
	"status": "accepted",
	"token": "INSERT TOKEN HERE"
}
```  

###Sending tokens to the server
The tokens are transferred to the server in a *Token-header* (the header should be called 'Token' and should contain the JWT in plain-text)

###Information about tokens
The token is signed with the private key from the server. The very same private key that is used to encrypt the TLS-stream. 
This means that a verified certificate for TLS will give the client the possibility to verify the signing and key against a CA.

#Install go on linux
Installera go genom att skriva följande:
```
sudo apt-get update
sudo apt-get install golang
```
###Set GOPATH
Lägg till följande rader i slutet i filen `~/.bashrc`. Den kan öppnas med texteditorn nano genom `nano`.
```
# GO
export GOPATH=$HOME/Documents/work
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

Källa: `http://stackoverflow.com/questions/21001387/how-do-i-set-the-gopath-environment-variable-on-ubuntu-what-file-must-i-edit`

#Generera egen signerade certifikat
Skapa RSA public och private key genom (fyll i de första fälten, du kan hoppa över dem genom att klicka enter, välj ett dyligt lösenord):
```
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365
```
Detta skpar filerna

* cert.pem
* key.pem

Avkryptera (dvs. ta bort lösenordet) private keyn genom:
```
openssl rsa -in key.pem  -out key.unencrypted.pem
```
Detta skapar filen:

* key.unencrypted.pem
