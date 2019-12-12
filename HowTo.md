
[intro]
demo driven presentation talk about how to quickly obtain a poc, push it to prod and use the atlas open source service broker to attain end-to-end security.


[Vlad] Slide 3: of course use mongo, but not just mongo, MongoDB Atlas
[Vlad] Slide 4: We already know this
[Vlad] Slide: 5 -- We go off deploy an atas cluster and get the connection String, Thanks Brad for covering the costs :)

[Vlad] Slide 6-7, So we've got the database now I need the application. Of c we're at springone so we use spring!
[Vlad] Slike 8 - Push to prod, the cloud sure no problema amigo! `cf push`

....
....
[Diana] Slide 9 -> Did you just go to prod?? hold on, hold on, hold on....
  The dreaded checklist of sec reqs:
    - tls/ssl/security passwords exposed, who has access/ rotation? valid period
[Diana] Slide 10 --> Well, dear security concerned peeps , let me introduce you a service broker that spins off atlas cluster and by the way securely stores those pesky credentials, yes the ones  you've just exposed in your environemnt, given to N developers (interns) which we all know translate to a plain text sticky notes on your monitor 

  - What is OSB? industry standard way to spin off services. Composed of 6 or 7 standard API.

[Diana] slide 10 - explain the arch.
[diana] slide 11 demo time

// The admin
  configure manifest
  How? Your admin pushes the atlas service broker app
  promotes it to a service
  enables plan that make sense for the company

// The dev
creates a service
pushes the app
binds the service to the service
TADA! Look ma no hands.

POC is now PRod ready!

[Diana] prod time demo


All right, so in the past hour we've shown how to 

Build a full blown REST API using Spring  and MongoDB Atlas -> Done, . We then pushed that application from my local environment to a prod-like environemtn which does not expose credentials to the developers.



This initial effort on the Atlas OSB was put together as a fun pet project for SpringOne in DC. The project has since gained traction. It's now being productized so we're working with PM/Cloud/Tech SUpport. Furthermore, because this adheres to OSB it can be easily ported in K8s if dockerized/helcharts. As of right now, we do not have an official team working on this and that will likely not be the case til 4.2 comes out so if you have spare cycles and would like get your hands dirty with go please reachout.



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
