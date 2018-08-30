pipeline {
  agent {
    dockerfile {
      filename 'DOCKERFILE'
    }

  }
  stages {
    stage('Initialise') {
      steps {
        echo 'Initialise'
      }
    }
    stage('Build') {
      steps {
        sh '''xgo --targets=linux/amd64,linux/386,windows/amd64,windows/386,darwin/amd64,darwin/386 github.com/republicprotocol/renex-swapper-go/cmd/swapper
'''
      }
    }
  }
}