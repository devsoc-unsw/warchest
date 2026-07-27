# *Warchest Money Module*
## What is Money Module:
This is an implementation of a logged ledger around Tigerbeetle as the OLTP and Postgres as the OLGP.


# For Developers:
## Use of directories:
### Cmd:
Money module drivers and handlers here 
### Internal:
Private code, for money-module this will just be our config
Folder has special treatment from Go compilers
Money-Module business logic should also exist here
Simple MVC like structure with grpc/service/repository layers here
### Pkg:
Public code, basically library code/generic helps should be placed here
### Cmd:
Only main.go will live here for now
## Endpoints:
### Example Endpoint