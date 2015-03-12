# Introduction #

This assumes that openssl has been installed locally. This step is also not necessary if you use either tls = false or remove the certpath setting line in your config file (in which case a random cert will be generated on startup).


# Details #

Command sequence:

make a user-specific golem directory:
```
mkdir ~/.golem```

make a PEM encoded key
```
openssl genrsa -out ~/.golem/key.pem 1024```
generate the certificate request
```
openssl req -new -key ~/.golem/key.pem -out ~/.golem/certificate.csr```

generate the certificate file from the request using the key
```
openssl x509 -req -days 365 -in ~/.golem/certificate.csr -signkey ~/.golem/key.pem -out ~/.golem/certificate.crt```

translate the CRT certificate to DER format (1st step to generate a PEM)
```
openssl x509 -in ~/.golem/certificate.crt -out ~/.golem/certificate.der -outform DER```

translate the DER formate to a PEM file
```
openssl x509 -in ~/.golem/certificate.der -inform DER -out ~/.golem/certificate.pem -outform PEM```