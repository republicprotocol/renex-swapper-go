pipeline {
  agent {
    docker {
      image 'golang:alpine'
    }

  }
  stages {
    stage('Initialise') {
      steps {
        sh '''go get -u github.com/karalabe/xgo
'''
      }
    }
    stage('Build') {
      steps {
        sh 'xgo --targets=linux/amd64,linux/386,windows/amd64,windows/386,darwin/amd64,darwin/386 github.com/republicprotocol/renex-swapper-go/cmd/swapper'
      }
    }
  }
}