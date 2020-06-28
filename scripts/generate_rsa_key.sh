#!/bin/bash

PASS=${1}
openssl genrsa -passout pass:$PASS -out private.key 1024
openssl rsa -in private.key -outform PEM -passin pass:$PASS -pubout -out public.pem

mv private.key config/key/
mv public.pem config/key/