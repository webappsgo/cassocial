pipeline {
    agent { docker { image 'golang:alpine' } }

    options {
        disableConcurrentBuilds()
        timeout(time: 30, unit: 'MINUTES')
    }

    environment {
        PROJECTNAME = 'cassocial'
        PROJECTORG  = 'casapps'
        CGO_ENABLED = '0'
    }

    stages {
        stage('Build') {
            steps {
                sh '''
                    apk add --no-cache git

                    VERSION=$(cat release.txt 2>/dev/null || echo "0.0.0-dev")
                    BUILD_DATE=$(date +"%a %b %d, %Y at %H:%M:%S %Z")
                    COMMIT_ID=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
                    if [ -f site.txt ]; then
                      OFFICIALSITE=$(cat site.txt)
                    else
                      OFFICIALSITE=${OFFICIALSITE:-}
                    fi

                    LDFLAGS="-s -w -X 'main.Version=${VERSION}' -X 'main.CommitID=${COMMIT_ID}' -X 'main.BuildDate=${BUILD_DATE}' -X 'main.OfficialSite=${OFFICIALSITE}'"

                    mkdir -p binaries

                    for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 freebsd/amd64 freebsd/arm64; do
                        OS=${platform%/*}
                        ARCH=${platform#*/}
                        OUTPUT=binaries/${PROJECTNAME}-${OS}-${ARCH}
                        [ "$OS" = "windows" ] && OUTPUT=${OUTPUT}.exe

                        echo "Building ${OS}/${ARCH}..."
                        GOOS=${OS} GOARCH=${ARCH} CGO_ENABLED=0 go build -buildvcs=false \
                            -ldflags "${LDFLAGS}" -o "${OUTPUT}" ./src

                        CLI_OUTPUT=binaries/${PROJECTNAME}-cli-${OS}-${ARCH}
                        [ "$OS" = "windows" ] && CLI_OUTPUT=${CLI_OUTPUT}.exe

                        GOOS=${OS} GOARCH=${ARCH} CGO_ENABLED=0 go build -buildvcs=false \
                            -ldflags "${LDFLAGS}" -o "${CLI_OUTPUT}" ./src/client
                    done
                '''
            }
        }

        stage('Test') {
            steps {
                sh '''
                    apk add --no-cache git
                    go vet ./...
                    go test -v -coverprofile=coverage.out -covermode=atomic ./...
                    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '"'"'{print $3}'"'"' | tr -d '"'"'%'"'"')
                    echo "Total coverage: ${COVERAGE}%"
                    if awk "BEGIN{exit !(${COVERAGE} < 100)}"; then
                        echo "Coverage ${COVERAGE}% is below the required 100% threshold"
                        exit 1
                    fi
                '''
            }
        }

        stage('Security') {
            parallel {
                stage('Secret Scan') {
                    steps {
                        sh 'docker run --rm -v "$(pwd):/repo" trufflesecurity/trufflehog:latest git file:///repo --since-commit HEAD~1 --only-verified --fail'
                    }
                }
                stage('Vuln Scan') {
                    steps {
                        sh '''
                            apk add --no-cache git
                            go install golang.org/x/vuln/cmd/govulncheck@latest
                            govulncheck ./...
                        '''
                    }
                }
            }
        }

        stage('Release') {
            when { tag 'v*' }
            steps {
                sh '''
                    apk add --no-cache git curl

                    VERSION=$(cat release.txt 2>/dev/null || echo "${TAG_NAME#v}")
                    cd binaries
                    sha256sum * > sha256sums.txt

                    curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
                    cd ..
                    syft . -o cyclonedx-json > binaries/sbom.cyclonedx.json

                    echo "${VERSION}" > binaries/version.txt
                '''
            }
            post {
                success {
                    archiveArtifacts artifacts: 'binaries/**', fingerprint: true
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
