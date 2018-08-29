pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        git(url: 'https://github.com/republicprotocol/renex-swapper-go.git', branch: 'master')
        sh '''dep ensure
go build ./cmd/swapper/swapper.go
go build ./cmd/installer/installer.go
mv swapper ~/builds/swapper_ubuntu
mv installer ~/builds/installer_ubuntu

'''
      }
    }
  }
}