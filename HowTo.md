
 # How To

## Login to Env

```bash
➜ cf login -a https://api.sys.pas.pcf.cloud.oskoss.com --skip-ssl-validation
  admin
  https://opsman.pcf.cloud.oskoss.com/api/v0/deployed/products/cf-ee61061cbfd3c03073c1/credentials/.uaa.admin_credentials 
```

## Create Org/Space

```bash
cf create-org atlas
cf target -o "atlas"
cf create-space broker
cf target -s broker
```

## See Apps

```bash
➜  cf apps
Getting apps in org atlas / space broker as admin...
OK
name           requested state   instances   memory   disk   urls
atlas-broker   started           1/1         1G       1G     atlas-broker.apps.pas.pcf.cloud.oskoss.com
```

## Building the Go App

```bash

cd /Users/d/go/src/code.cloudfoundry.org/credhub-cli
git checkout 863774d30866c04d1dacb75df13ecfbe5f84d163

cd /Users/d/go/src/github.com/desteves/mongodb-atlas-service-broker
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
mkdir bin
mv mongodb-atlas-service-broker bin
cd bin
```

## Pushing the Go App

```bash
➜  cf buildpacks
➜  bin git:(master) ✗ pwd
/Users/d/go/src/github.com/desteves/mongodb-atlas-service-broker/bin

➜  cat manifest.yml
applications:
- name: atlas-broker
  command: ./mongodb-atlas-service-broker
  buildpack: binary_buildpack
  env:
      ATLAS_USERNAME: diana.esteves
      ATLAS_API_KEY: d76d39bf-b356-4e8d-bbd0-b78998e75e56
      ATLAS_GROUP_ID: 5b75e84b3b34b9469d01b20e
      ATLAS_HOST: cloud.mongodb.com
      UAA_ADMIN_CLIENT_SECRET: W6q3YNrRbUrx2-H8wd5sLRTTVRCRhRqH
      SECURITY_USER_NAME: admin
      SECURITY_USER_PASSWORD: admin

➜  cf push  -f manifest.yml
➜  cf logs atlas-broker --recent
➜  cf logs atlas-broker

```

## Create Service Broker

```bash
cf service-brokers
cf create-service-broker mongodb-atlas-service-broker admin admin http://atlas-broker.apps.pas.pcf.cloud.oskoss.com
cf service-access
cf enable-service-access atlas
cf marketplace
```

## Create Service

`cf create-service atlas aws_dev my-frist-atlas-sb`

## Bind Service

```bash
cf apps
cf bind-service my-app my-frist-atlas-sb
cf push  -f manifest.yml
cf bind-service my-app my-frist-atlas-sb
cf env my-app
  
```

## Deploy Spring App

## Bind Spring App
➜  bin git:(master) ✗ cf bind-service springdemo atlas2
Binding service atlas2 to app springdemo in org atlas / space broker as admin...
OK
TIP: Use 'cf restage springdemo' to ensure your env variable changes take effect
➜  bin git:(master) ✗ cf restage

<!-- 
 1018  cf create-service atlas aws_dev atlas2
 1020  cf bind-service atlas-broker atlas2
 1021* cf logs atlas-broker
 1022* cf logs atlas2
 1023* cf logs atlas-broker
 1024  cf ssh atlas-broker
 1025  cf run-task atlas-broker env
 1026  cf bind-service atlas-broker atlas2
 1045  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
 1046  mv mongodb-atlas-service-broker bin
 1047  cd bin
 1048  cf push  -f manifest.yml
 1049  cf bind-service atlas-broker atlas2 -->
