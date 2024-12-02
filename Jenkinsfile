def DeployRegion(regionName, dbHost, dbName, branch, APP_DIR ) {
    try {             
            sh "psql -U postgres -d ${dbName} -h ${dbHost} < database/up.sql"
            echo "Moving files to Application Directory"
            sh "supervisorctl stop ${regionName} && cp  server  ${APP_DIR}"
            echo "Starting service"
            sh "supervisorctl start ${regionName} "
    }
    catch (Exception e) {
        echo "Deployment failed for ${regionName}, starting rollback, RESON: ${e}"
        rollbackRegion(APP_DIR, regionName, branch )
        error "Deployment for ${regionName} failed. Rolled back successfully."
    }
}


def rollbackRegion(APP_DIR, regionName, branch) {
    echo "Restoring previous version for ${regionName}..."
    sh "supervisorctl stop ${regionName} && rm  ${APP_DIR}/server && cp  server ${APP_DIR} && supervisorctl start ${regionName}"
}
    
pipeline {      
    agent {
        label (env.BRANCH_NAME == 'dev' ? 'backend_beta' : 'backend_app')
    }
    environment {
        branch = sh(returnStdout: true, script: 'git branch --show-current').trim()
        // Default directory for deployment (main branch)
        APP_DIR = ""
        BACKUP_DIR = ""
    
    }
    options {
        // Skip waiting for input at other stages (assuming this applies to all stages)
        disableConcurrentBuilds(abortPrevious: true)
    }

    stages {
        stage('Setup') {
            steps {
                script {

                    if (env.BRANCH_NAME == 'dev') {
                        APP_DIR = "/var/www/beta/server"
                    } else if (env.BRANCH_NAME == 'stable') {
                        def regions = ['ahal', 'mary', 'balkan', 'lebap', 'dashoguz']
                        for (region in regions) {
                            APP_DIR = "/var/www/regions/${region}/server"
                        }
                    } else if (env.BRANCH_NAME == 'main') {
                        APP_DIR = "/var/www/server"
                    } else {
                        error("CANNOT FIND KNOWN BRANCH TO DEPLOY")
                    }
                    // Set BACKUP_DIR based on the branch
                    if (env.BRANCH_NAME == 'main' || env.BRANCH_NAME == 'dev') {                      
                        BACKUP_DIR = '/home/sohbet/codes/server'
                    } else {
                        BACKUP_DIR = '/home/sohbet/codes/region/server'
                    }

                    // Set the environment variables for use in other stages
                    env.APP_DIR = APP_DIR
                    env.BACKUP_DIR = BACKUP_DIR

                    // Print values to ensure they are set correctly
                    echo "BACKUP_DIR is set to ${BACKUP_DIR}"
                    echo "APP_DIR is set to ${APP_DIR}"
                }
            }
        }

        stage('Backup') {
            steps {
                script {
                        try {
                            sh "rm ${BACKUP_DIR}/server && cp ${APP_DIR}/server ${BACKUP_DIR}"
                        } catch(Exception e) {
                            error('Don\'t Panic! Everything works fine. Pipeline failed due to a Backup failure for rollback.')
                        }
                    
                }
            }
        }

        stage("Checkout"){
            steps{
                git branch: "${env.BRANCH_NAME}", credentialsId: 'ci-mekdep', url: 'https://github.com/mekdep/server'
            }
        }
        
        stage("Build"){
            steps{
                script {
                    sh "go build ."
                }
                
            }
        }
        
        stage("Stop Monitoring Service"){
            steps{
                script{
                    switch(env.BRANCH_NAME) {
                        case "dev":
                            sshagent(credentials: ['sohbet_jenkins']) {
                                sh "supervisorctl stop monitor_api"
                            }
                        case "stable":
                            sshagent(credentials: ['sohbet_jenkins']) {
                                sh "ssh -p 9021 sohbet@192.168.1.101 'supervisorctl stop monitor_api'"
                                sh "ssh -p 9021 sohbet@192.168.1.110 'supervisorctl stop monitor-api'"
                            }
                        case "main":
                            sshagent(credentials: ['sohbet_jenkins']) {
                                sh "ssh -p 9021 sohbet@192.168.1.101 'supervisorctl stop monitor_api'"
                                sh "ssh -p 9021 sohbet@192.168.1.110 'supervisorctl stop monitor-api'"
                            }
                        }
                    }
                }
        }
        stage("Deploy"){
            steps{
                script{
                    if (env.BRANCH_NAME == 'stable') {
                        // Define region-specific database hosts and names
                        def regionDBConfig = [
                            'ahal'   : [dbHost: '192.168.1.102', dbName: 'ahal_db', APP_DIR: '/var/www/regions/ahal/server'],
                            'mary'   : [dbHost: '192.168.1.106', dbName: 'mary_db', APP_DIR: '/var/www/regions/mary/server'],
                            'balkan'   : [dbHost: '192.168.1.107', dbName: 'balkan_db', APP_DIR: '/var/www/regions/balkan/server'],
                            'lebap'   : [dbHost: '192.168.1.108', dbName: 'lebap_db', APP_DIR: '/var/www/regions/lebap/server'],
                            'dashoguz'   : [dbHost: '192.168.1.105', dbName: 'dashoguz_db', APP_DIR: '/var/www/regions/dashoguz/server']
                        ]
                        
                        // Loop through each region and deploy
                        for (region in regionDBConfig.keySet()) {
                            def dbHost = regionDBConfig[region].dbHost
                            def dbName = regionDBConfig[region].dbName
                            APP_DIR = regionDBConfig[region].APP_DIR
                            DeployRegion(region, dbHost, dbName, env.BRANCH_NAME, APP_DIR)
                        }
                    } else if (env.BRANCH_NAME == 'dev' || env.BRANCH_NAME == 'main'){
                        def dbName = (env.BRANCH_NAME == 'dev') ? 'mekdep_beta' : 'mekdep_db'
                        def dbHost = (env.BRANCH_NAME == 'dev') ? '192.168.1.107' : '192.168.1.112'
                        def region = (env.BRANCH_NAME == 'dev') ? 'beta' : 'app'
                        DeployRegion(region, dbHost, dbName, env.BRANCH_NAME, APP_DIR)
                    }else {
                        error("Couldnt found known branch to Deploy")
                    }
                }
            }
        }
    }

}
