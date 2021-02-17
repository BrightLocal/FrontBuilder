#!/usr/bin/env groovy
pipeline {
    agent any

    options {
        timeout(time: 10, unit: 'MINUTES')
    }

    environment {
        DOCKER_IMAGE          = 'hub.bll-i.co.uk/ci/golang:latest'
        PROJECT_NAME          = 'FrontBuilder'
        DOCKER_CONTAINER_NAME = "${PROJECT_NAME}_${GIT_COMMIT.take(8)}_${BUILD_NUMBER}"
        PROJECT_PATH          = "/go/src/github.com/BrightLocal/${PROJECT_NAME}"
        COVERAGE_BADGE_SH     = "curl --data-urlencode -s \"https://img.shields.io/badge/Test%20coverage-\$(go tool cover -func=${PROJECT_PATH}/coverage.out | tail -1 | awk \'{print \$3}\')25-brightgreen.svg?longCache=true&style=flat\" > ${PROJECT_PATH}/coverage_badge.svg"
    }

    stages {
        stage('Prepare') {
            steps {
                echo 'Pulling newest image'
                sh "docker pull ${DOCKER_IMAGE}"
                echo 'Start container'
                sh """
                   docker run -d --rm --net=host --name ${DOCKER_CONTAINER_NAME} \\
                   -v /home/jenkins/.ssh:/home/jenkins/.ssh \\
                   -v /home/jenkins/.cache:/home/jenkins/.cache \\
                   -v `pwd`:${PROJECT_PATH} \\
                   ${DOCKER_IMAGE}
                   """
            }
        }
        stage('Test') {
            steps {
                echo 'Running go tests'
                sh "docker exec -i ${DOCKER_CONTAINER_NAME} bash -c 'cd ${PROJECT_PATH} && go test -coverprofile=coverage.out ./... -timeout 60s'"
            }
        }
        stage('Creating coverage shell script') {
            when {
                expression { GIT_BRANCH == 'main' }
            }
            steps {
                echo 'Create script'
                writeFile file: 'coverage.sh', text: COVERAGE_BADGE_SH
            }
        }
        stage('Copy coverage shell script') {
            when {
                expression { GIT_BRANCH == 'main' }
            }
            steps {
                echo 'Copy script'
                sh "docker cp coverage.sh ${DOCKER_CONTAINER_NAME}:${PROJECT_PATH}/coverage.sh"
            }
        }
        stage('Run sh') {
            when {
                expression { GIT_BRANCH == 'main' }
            }
            steps {
                echo 'Run sh script'
                sh "docker exec -i ${DOCKER_CONTAINER_NAME} bash -c 'sh ${PROJECT_PATH}/coverage.sh'"
            }
        }
        stage('Deploy badge') {
            when {
                expression { GIT_BRANCH == 'main' }
            }
            steps {
                sshagent (credentials: ['deploy-as-sites-key']) {
                    sh "scp -B -q -o StrictHostKeyChecking=no -r coverage_badge.svg sites@91.196.148.90:/var/www/html/coverage/badges/${PROJECT_NAME}"
                }
            }
        }
    }

    post {
        always {
            echo 'Removing docker container'
            sh "docker stop ${DOCKER_CONTAINER_NAME} || true"
        }
    }
}
