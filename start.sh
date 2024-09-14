#!/usr/bin/env bash

pushd coldmfa/app || exit 1
# In dev, run `npm run dev`
npm run build
popd || exit 1

# In dev, run `go run . dev`
go run .
