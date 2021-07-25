#!/usr/bin/env sh

protoc --go_out=. --go_opt=paths=source_relative *.proto;
