# insights-client

## DESCRIPTION

**insights-client** is designed to help customers resolve issues affecting business operations in Red Hat Enterprise Linux and hybrid cloud infrastructure environments. The client manages data collection and upload to Red Hat Insights. Learn more at https://console.redhat.com/insights/. 


## USE-CASES

1. Register the host:
   `insights-client --register`
   
   Register the host while setting Insights Inventory display name:
   `insights-client --register --display-name [HOSTNAME]`

2. Unregister the host.
   `insights-client --unregister`
 
3. Change the Insights Inventory display name of a registered host:
   `insights-client --display-name [HOSTNAME]`

4. Display registration status of the host:
   `insights-client --status`

5. Perform data collection and upload on a registered host:
   `insights-client `

6. Perform data collection for Compliance on a registered host:
   `insights-client -m compliance`

7. Perform data collection and keep it uncompressed without uploading:
   `insights-clent --output-dir /var/cache/insights-client/archive/`

8. Perform data collection and keep it compressed without uploading:
   `insights-client --output-file /var/cache/insights-client/archive.tar.xz`

9. Upload previously collected archive on a registered host:
   `insights-client --payload /var/cache/insights-client/archive.tar.xz --content-type application/...+txz`


### COMMON OPTIONS

- `--debug`
  Display colored log information to standard error instead of printing them to the log file.
- `--format [FORMAT]`
  Present output in a human- or machine-readable format. Machine-readable format may not be available for all commands.


## FILES

- `/etc/insights-client/insights-client.conf`
  The configuration file for the client.
- `/etc/insights-client/insights-client.conf.d/*.conf`
  User-specific configurations loaded on top of the configuration file. See `insights-client.conf(5)` for more details.
- `/var/log/insights-client/insights-client.log`
  The log file.
