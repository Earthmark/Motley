@echo off

echo pushd client
pushd client
echo yarn install
call yarn install
echo yarn build
call yarn build
echo popd
popd
echo pushd server
pushd server
echo go generate ./...
go generate ./...
echo go build
go build
popd
