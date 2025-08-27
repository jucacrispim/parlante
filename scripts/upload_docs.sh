#!/bin/bash

cd docs/build
mv html parlante
tar -czf docs.tar.gz parlante

curl -F 'file=@docs.tar.gz' https://docs.poraodojuca.dev/e/ -H "Authorization: Key $TUPI_AUTH_KEY"
