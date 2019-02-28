/*

goal of the service is to help developers trust that they have backups
use some machine learning to predict backup size and report when abnormal

For who?

people like me(jackdoe) that have few machines and few databases and want
to know that their backups actually work

or some of my friends who always call me with their broken backups

Pricing

  * free [ this is how i would use it ]
    start the daemon
    configure it and you are good to go (setup sms, email and mysql)
    you can also make it upload to your s3 bucket
    create tokens for backup clients

    example flow:
      on backup server:
          install baxx
          baxx -conf /etc/baxx.conf

      the baxx daemon needs sql (pg, mysql, sqlite) to store the metadata


    on other servers you simply do
      mysqldump | [encrypt -p passfile] | curl -' -k https://baxIP.IP.IP.IP/v1/upload/$CLIENT/$TOKEN/mysql.gz
      (encrypt is optional, and you might want to ignore it, you might also want to have SSL properly)

    another example upload everything, only different files will be added
     find . -type f -exec curl --binary-data @{} https://baxx.dev/v1/upload/$CLIENT/$TOKEN/{} \;

     for i in $(find . -type f); do
        curl -f https://baxx.dev/v1/diff/$CLIENT/$TOKEN/$(shasum $i | cut -f 1 -d ' ') \
        && curl --binary-data @$i https://baxx.dev/v1/upload/$CLIENT/$TOKEN
     done

    FIXME: find more efficient 1 liner

    - notifications
        + per directory
          . schecule [ when no new files are added in N hours ]
          . size [ when size does not change in N hours ]
          . when new files are smaller than old files




  api:
    create client
    create token for client
    upload file in client/token
    list files in directory in client/token
    set config for token


*/

/*
TODO:
   * add sha check endpoint
   * encrypt everything with a salt [done]
   * compress [done]
   * send notifications
   * write only tokens
   * archive configutragion
*/
