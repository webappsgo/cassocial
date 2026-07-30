pipeline {
    agent none

    triggers {
        // Daily build at 3am UTC (matches GitHub Actions daily.yml cadence)
        cron('0 3 * * *')
    }

    environment {
        PROJECT_NAME = 'cassocial'
        PROJECT_ORG = 'casapps'
        BINDIR = 'binaries'
        RELDIR = 'releases'
        // Go cache bind-mounted from host: GO_CACHE (mod) and GO_BUILD (build cache)

        // ----- GITHUB -----
        GIT_FQDN = 'github.com'
        // Jenkins credentials ID
        GIT_TOKEN = credentials('github-token')
        REGISTRY = "ghcr.io/${PROJECT_ORG}/${PROJECT_NAME}"
    }

    stages {
        stage('Setup') {
            agent { label 'amd64' }
            steps {
                script {
                    // Determine build type and version
                    if (env.TAG_NAME) {
                        // Release build (tag push) - matches release.yml
                        env.BUILD_TYPE = 'release'
                        env.VERSION = env.TAG_NAME.replaceFirst('^v', '')
                    } else if (env.BRANCH_NAME == 'beta') {
                        // Beta build - matches beta.yml
                        env.BUILD_TYPE = 'beta'
                        env.VERSION = sh(script: 'date -u +"%Y%m%d%H%M%S"', returnStdout: true).trim() + '-beta'
                    } else if (env.BRANCH_NAME == 'main' || env.BRANCH_NAME == 'master') {
                        // Daily build - matches daily.yml
                        env.BUILD_TYPE = 'daily'
                        env.VERSION = sh(script: 'date -u +"%Y%m%d%H%M%S"', returnStdout: true).trim()
                    } else {
                        // Other branches - dev build
                        env.BUILD_TYPE = 'dev'
                        env.VERSION = sh(script: 'date -u +"%Y%m%d%H%M%S"', returnStdout: true).trim() + '-dev'
                    }
                    env.COMMIT_ID = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.BUILD_DATE = sh(script: 'date +"%a %b %d, %Y at %H:%M:%S %Z"', returnStdout: true).trim()
                    // OFFICIAL_SITE (optional): site.txt wins; otherwise use Jenkins credentials or leave empty
                    // Never guess or assume - must be explicitly defined by user
                    env.OFFICIAL_SITE = sh(script: '[ -f site.txt ] && cat site.txt || echo "${OFFICIAL_SITE:-}"', returnStdout: true).trim()
                    env.LDFLAGS = "-s -w -X 'main.Version=${env.VERSION}' -X 'main.CommitID=${env.COMMIT_ID}' -X 'main.BuildDate=${env.BUILD_DATE}' -X 'main.OfficialSite=${env.OFFICIAL_SITE}'"
                    env.HAS_CLI = sh(script: '[ -d src/client ] && echo true || echo false', returnStdout: true).trim()
                    env.HAS_AGENT = sh(script: '[ -d src/agent ] && echo true || echo false', returnStdout: true).trim()
                }
                sh 'mkdir -p ${BINDIR} ${RELDIR}'
                echo "Build type: ${BUILD_TYPE}, Version: ${VERSION}"
            }
        }

        stage('Security') {
            agent { label 'amd64' }
            parallel {
                stage('Secret Scan') {
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/repo \
                                -w /repo \
                                trufflesecurity/trufflehog:latest \
                                git file:///repo --since-commit "${GIT_PREVIOUS_SUCCESSFUL_COMMIT:-HEAD~1}" --only-verified --fail
                        '''
                    }
                }
                stage('Vuln Scan') {
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                casjaysdev/go:latest \
                                govulncheck ./src/...
                        '''
                    }
                }
            }
        }

        stage('Build Server') {
            parallel {
                stage('Linux AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=linux \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-linux-amd64 ./src
                        '''
                    }
                }
                stage('Linux ARM64') {
                    agent { label 'arm64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=linux \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-linux-arm64 ./src
                        '''
                    }
                }
                stage('Darwin AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=darwin \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-darwin-amd64 ./src
                        '''
                    }
                }
                stage('Darwin ARM64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=darwin \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-darwin-arm64 ./src
                        '''
                    }
                }
                stage('Windows AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=windows \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-windows-amd64.exe ./src
                        '''
                    }
                }
                stage('Windows ARM64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=windows \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-windows-arm64.exe ./src
                        '''
                    }
                }
                stage('FreeBSD AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=freebsd \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-freebsd-amd64 ./src
                        '''
                    }
                }
                stage('FreeBSD ARM64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=freebsd \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-freebsd-arm64 ./src
                        '''
                    }
                }
            }
        }

        // CLI builds - only if src/client/ exists (matches GitHub Actions)
        stage('Build CLI') {
            when {
                expression { env.HAS_CLI == 'true' }
            }
            parallel {
                stage('CLI Linux AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=linux \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-linux-amd64 ./src/client
                        '''
                    }
                }
                stage('CLI Linux ARM64') {
                    agent { label 'arm64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=linux \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-linux-arm64 ./src/client
                        '''
                    }
                }
                stage('CLI Darwin AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=darwin \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-darwin-amd64 ./src/client
                        '''
                    }
                }
                stage('CLI Darwin ARM64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=darwin \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-darwin-arm64 ./src/client
                        '''
                    }
                }
                stage('CLI Windows AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=windows \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-windows-amd64.exe ./src/client
                        '''
                    }
                }
                stage('CLI Windows ARM64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=windows \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-windows-arm64.exe ./src/client
                        '''
                    }
                }
                stage('CLI FreeBSD AMD64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=freebsd \
                                -e GOARCH=amd64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-freebsd-amd64 ./src/client
                        '''
                    }
                }
                stage('CLI FreeBSD ARM64') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v ${WORKSPACE}:/app \
                                -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                                -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                                -w /app \
                                -e CGO_ENABLED=0 \
                                -e GOOS=freebsd \
                                -e GOARCH=arm64 \
                                casjaysdev/go:latest \
                                go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-cli-freebsd-arm64 ./src/client
                        '''
                    }
                }
            }
        }

        // Agent builds - only if src/agent/ exists (not present in this repo; stage is a no-op via `when`)
        stage('Build Agent') {
            when {
                expression { env.HAS_AGENT == 'true' }
            }
            steps {
                sh '''
                    docker run --rm \
                        --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                        -v ${WORKSPACE}:/app \
                        -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                        -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                        -w /app \
                        -e CGO_ENABLED=0 \
                        casjaysdev/go:latest \
                        go build -buildvcs=false -trimpath -ldflags "${LDFLAGS}" -o ${BINDIR}/${PROJECT_NAME}-agent ./src/agent
                '''
            }
        }

        stage('Test') {
            agent { label 'amd64' }
            steps {
                sh '''
                    docker run --rm \
                        --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                        -v ${WORKSPACE}:/app \
                        -v ${GO_CACHE:-$HOME/go/pkg/mod}:/usr/local/share/go/pkg/mod \
                        -v ${GO_BUILD:-$HOME/.cache/go-build/${PROJECT_NAME}}:/usr/local/share/go/cache \
                        -w /app \
                        -e CGO_ENABLED=0 \
                        -e GOFLAGS=-buildvcs=false \
                        casjaysdev/go:latest \
                        sh -c "go test -v -coverprofile=/tmp/coverage.out -covermode=atomic ./src/... && PCT=\\$(go tool cover -func=/tmp/coverage.out | awk '/^total:/ {gsub(\\"%\\",\\"\\",\\$3); print int(\\$3)}') && echo Coverage: \\${PCT}% && [ \\"\\$PCT\\" -ge 60 ]"
                '''
            }
        }

        // Stable Release - matches release.yml (tag push only)
        stage('Release: Stable') {
            agent { label 'amd64' }
            when {
                expression { env.BUILD_TYPE == 'release' }
            }
            steps {
                sh '''
                    echo "${VERSION}" > ${RELDIR}/version.txt

                    for f in ${BINDIR}/${PROJECT_NAME}-*; do
                        [ -f "$f" ] || continue
                        cp "$f" ${RELDIR}/
                    done

                    cd ${RELDIR} && sha256sum ${PROJECT_NAME}-* > sha256sums.txt && cd ..

                    tar --exclude='.git' --exclude='.github' --exclude='.gitea' \
                        --exclude='.forgejo' --exclude='binaries' --exclude='releases' \
                        --exclude='*.tar.gz' \
                        -czf ${RELDIR}/${PROJECT_NAME}-${VERSION}-source.tar.gz .
                '''
                archiveArtifacts artifacts: 'releases/*', fingerprint: true
            }
        }

        // Beta Release - matches beta.yml (beta branch only)
        stage('Release: Beta') {
            agent { label 'amd64' }
            when {
                expression { env.BUILD_TYPE == 'beta' }
            }
            steps {
                sh '''
                    echo "${VERSION}" > ${RELDIR}/version.txt

                    for f in ${BINDIR}/${PROJECT_NAME}-*; do
                        [ -f "$f" ] || continue
                        cp "$f" ${RELDIR}/
                    done
                '''
                archiveArtifacts artifacts: 'releases/*', fingerprint: true
            }
        }

        // Daily Release - matches daily.yml (main/master + scheduled)
        stage('Release: Daily') {
            agent { label 'amd64' }
            when {
                expression { env.BUILD_TYPE == 'daily' }
            }
            steps {
                sh '''
                    echo "${VERSION}" > ${RELDIR}/version.txt

                    for f in ${BINDIR}/${PROJECT_NAME}-*; do
                        [ -f "$f" ] || continue
                        cp "$f" ${RELDIR}/
                    done
                '''
                archiveArtifacts artifacts: 'releases/*', fingerprint: true
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
